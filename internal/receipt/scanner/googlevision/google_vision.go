package googlevision

import (
	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/receipt/scanner"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	vision "cloud.google.com/go/vision/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type GoogleVisionScanner struct {
	client *vision.ImageAnnotatorClient
	apiKey string
}

func MustInit() *GoogleVisionScanner {
	cfg := conf.Conf()
	if cfg.OCRService != "googlevision" {
		panic(errs.ErrGVisionServiceInit)
	}

	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithAPIKey(cfg.GoogleVisionConfig.APIKey))
	if err != nil {
		panic(errors.Join(errs.ErrGVisionAPI, err))
	}

	return &GoogleVisionScanner{
		client: client,
		apiKey: cfg.GoogleVisionConfig.APIKey,
	}
}

func (g *GoogleVisionScanner) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

func (g *GoogleVisionScanner) ScanReceipt(imagePath string) (*scanner.ReceiptData, error) {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, errors.Join(errs.ErrReceiptImageNotFound, err)
	}

	logs.Log().Info("Scanning receipt image", zap.String("path", imagePath))

	data, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, errors.Join(errs.ErrReceiptInvalidImage, errs.ErrFileRead, err)
	}

	ctx := context.Background()
	image := &visionpb.Image{Content: data}
	annotation, err := g.client.DetectDocumentText(ctx, image, nil)
	if err != nil {
		return nil, errors.Join(errs.ErrGVisionAPI, err)
	}

	if annotation == nil || annotation.Text == "" {
		return nil, errs.ErrReceiptNoTextFound
	}

	logs.Log().Info("Text extracted from receipt", zap.Int("length", len(annotation.Text)))

	receiptData := g.parseReceiptText(annotation.Text)
	receiptData.RawText = annotation.Text
	return receiptData, nil
}

func (g *GoogleVisionScanner) parseReceiptText(text string) *scanner.ReceiptData {
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		fmt.Println("DEBUG", i, line)
	}

	data := &scanner.ReceiptData{
		Items:    []scanner.LineItem{},
		Currency: "PHP",
	}

	parseMerchant(lines, data)

	merchantEndIdx, _ := findSectionBounds(lines, "SALES", "SOLD TO")
	if merchantEndIdx != -1 {
		for i := merchantEndIdx; i < len(lines) && i < merchantEndIdx+5; i++ {
			upperLine := strings.ToUpper(strings.TrimSpace(lines[i]))
			if strings.Contains(upperLine, "SALES") {
				data.ReceiptType = "SALES INVOICE"
			} else if strings.Contains(upperLine, "NO") && i+1 < len(lines) {
				data.ReceiptNumber = strings.ToUpper(strings.TrimSpace(lines[i+1]))
			}
		}
	}
	parseCustomer(lines, data)
	parseLineItems(lines, data)
	parseTotalSales(lines, data)

	return data
}

func parseMerchant(lines []string, data *scanner.ReceiptData) {
	startIdx, endIdx := findSectionBounds(lines, "C-CHOICE", "SALES")
	if startIdx == -1 {
		return
	}
	if endIdx == -1 {
		endIdx = min(len(lines), startIdx+10)
	}

	for i := startIdx; i < endIdx; i++ {
		trimmedLine := strings.TrimSpace(lines[i])
		if trimmedLine == "" {
			continue
		}
		upperLine := strings.ToUpper(trimmedLine)

		switch {
		case strings.Contains(upperLine, "C-CHOICE"):
			data.MerchantName = upperLine
		case strings.Contains(upperLine, "BLK") || strings.Contains(upperLine, "LT") || strings.Contains(upperLine, "PHILIPPINES"):
			data.MerchantAddress = upperLine
		case strings.Contains(upperLine, "VAT") || strings.Contains(upperLine, "TIN"):
			if tin := extractValueAfterLabel(upperLine, "TIN"); tin != "" {
				data.MerchantTIN = formatTIN(tin)
			}
		case strings.Contains(upperLine, "PROP"):
			data.MerchantProp = upperLine
		}
	}
}

func parseCustomer(lines []string, data *scanner.ReceiptData) {
	startIdx, endIdx := findSectionBounds(lines, "SOLD TO", "ITEM DESCRIPTION")
	if startIdx == -1 {
		startIdx, endIdx = findSectionBounds(lines, "SOLD TO", "NATURE OF SERVICE")
	}
	if startIdx == -1 {
		return
	}

	if endIdx == -1 {
		endIdx = min(len(lines), startIdx+15)
	}

	for i := startIdx; i < endIdx; i++ {
		trimmedLine := strings.TrimSpace(lines[i])
		if trimmedLine == "" {
			continue
		}
		upperLine := strings.ToUpper(trimmedLine)

		switch {
		case strings.Contains(upperLine, "SOLD TO"):
			if i+1 < len(lines) {
				data.SoldTo = strings.ToUpper(strings.TrimSpace(lines[i+1]))
			}
		case strings.Contains(upperLine, "TIN"):
			if tin := extractValueAfterLabel(upperLine, "TIN"); tin != "" {
				data.CustomerTIN = formatTIN(tin)
			}
		case strings.Contains(upperLine, "ADDRESS"):
			preaddress := extractValueAfterLabel(upperLine, "ADDRESS")
			if preaddress != "" {
				preaddress += " "
			}
			if i+1 < len(lines) {
				data.CustomerAddress = preaddress + strings.ToUpper(strings.TrimSpace(lines[i+1]))
			} else {
				data.CustomerAddress = preaddress
			}
			data.CustomerAddress = strings.ReplaceAll(data.CustomerAddress, "SUBOY.", "SUBD.")
			data.CustomerAddress = strings.ReplaceAll(data.CustomerAddress, "BROY", "BRGY.")
		}
	}
}

func parseLineItems(lines []string, data *scanner.ReceiptData) {
	startIdx := -1
	for i, line := range lines {
		upperLine := strings.ToUpper(strings.TrimSpace(line))
		if strings.Contains(upperLine, "ITEM DESCRIPTION") || strings.Contains(upperLine, "NATURE OF SERVICE") {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		return
	}

	endIdx := -1
	for i := startIdx + 1; i < len(lines); i++ {
		upperLine := strings.ToUpper(strings.TrimSpace(lines[i]))
		if strings.Contains(upperLine, "SALES") && !strings.Contains(upperLine, "AMOUNT") {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		endIdx = len(lines)
	}

	itemDescIdx := startIdx
	quantityIdx := -1
	unitCostIdx := -1
	amountIdx := -1

	for i := startIdx; i < endIdx; i++ {
		upperLine := strings.ToUpper(strings.TrimSpace(lines[i]))
		switch {
		case strings.Contains(upperLine, "QUANTITY"):
			quantityIdx = i
		case strings.Contains(upperLine, "UNIT COST"):
			unitCostIdx = i
		case strings.Contains(upperLine, "AMOUNT") && !strings.Contains(upperLine, "TOTAL"):
			amountIdx = i
		}
	}

	if quantityIdx != -1 {
		for i := itemDescIdx + 1; i < quantityIdx; i++ {
			trimmedLine := strings.TrimSpace(lines[i])
			if trimmedLine != "" {
				upperLine := strings.ToUpper(trimmedLine)
				lineItem := scanner.LineItem{Name: upperLine}
				data.Items = append(data.Items, lineItem)
			}
		}
	}

	if quantityIdx != -1 && unitCostIdx != -1 {
		lineItemIdx := 0
		for i := quantityIdx + 1; i < unitCostIdx; i++ {
			if lineItemIdx >= len(data.Items) {
				break
			}
			trimmedLine := strings.TrimSpace(lines[i])
			if trimmedLine == "" {
				trimmedLine = "1"
			}
			data.Items[lineItemIdx].Quantity = strings.ToUpper(trimmedLine)
			lineItemIdx++
		}
		for lineItemIdx < len(data.Items) {
			data.Items[lineItemIdx].Quantity = "1"
			lineItemIdx++
		}
	}

	if unitCostIdx != -1 && amountIdx != -1 {
		lineItemIdx := 0
		i := unitCostIdx + 1
		for i < amountIdx && lineItemIdx < len(data.Items) {
			trimmedLine := strings.TrimSpace(lines[i])
			upperLine := strings.ToUpper(trimmedLine)

			if strings.Contains(upperLine, "PRICE") || strings.Contains(upperLine, "COST") {
				i++
				continue
			}

			if trimmedLine == "" {
				i++
				continue
			}

			combined, skipCount := combineNumbers(lines, i)
			data.Items[lineItemIdx].Price = combined
			lineItemIdx++
			i += 1 + skipCount
		}
	}

	if amountIdx != -1 {
		lineItemIdx := 0
		i := amountIdx + 1
		for i < endIdx && lineItemIdx < len(data.Items) {
			trimmedLine := strings.TrimSpace(lines[i])
			upperLine := strings.ToUpper(trimmedLine)

			if strings.Contains(upperLine, "SALES") || strings.Contains(upperLine, "TOTAL") {
				break
			}

			if trimmedLine == "" {
				i++
				continue
			}

			combined, skipCount := combineNumbers(lines, i)
			data.Items[lineItemIdx].Subtotal = combined
			lineItemIdx++
			i += 1 + skipCount
		}
	}
}

func parseTotalSales(lines []string, data *scanner.ReceiptData) {
	startIdx := -1
	for i, line := range lines {
		upperLine := strings.ToUpper(strings.TrimSpace(line))
		if strings.Contains(upperLine, "TOTAL SALES") && strings.Contains(upperLine, "VAT INCLUSIVE") {
			startIdx = i
			break
		}
	}

	if startIdx == -1 {
		return
	}

	endIdx := min(len(lines), startIdx+20)

	labelMap := make(map[string]int)
	valueQueue := []string{}

	for i := startIdx; i < endIdx; i++ {
		trimmedLine := strings.TrimSpace(lines[i])
		if trimmedLine == "" {
			continue
		}
		upperLine := strings.ToUpper(trimmedLine)

		switch {
		case strings.Contains(upperLine, "TOTAL SALES") && strings.Contains(upperLine, "VAT INCLUSIVE"):
			labelMap["VAT_INCLUSIVE"] = i
		case strings.Contains(upperLine, "LESS") && strings.Contains(upperLine, "VAT") && !strings.Contains(upperLine, "WITHHOLDING"):
			labelMap["LESS_VAT"] = i
		case strings.Contains(upperLine, "LESS") && strings.Contains(upperLine, "WITHHOLDING"):
			labelMap["LESS_WITHHOLDING"] = i
		case strings.Contains(upperLine, "AMOUNT NET OF VAT"):
			labelMap["NET_OF_VAT"] = i
		case strings.Contains(upperLine, "ADD") && strings.Contains(upperLine, "VAT"):
			labelMap["ADD_VAT"] = i
		case strings.Contains(upperLine, "TOTAL AMOUNT"):
			labelMap["TOTAL_AMOUNT"] = i
		case isNumericValue(trimmedLine):
			combined, skipCount := combineNumbers(lines, i)
			valueQueue = append(valueQueue, combined)
			i += skipCount
		}
	}

	type labelOrder struct {
		name  string
		index int
	}

	orderedLabels := make([]labelOrder, 0, len(labelMap))
	for name, idx := range labelMap {
		orderedLabels = append(orderedLabels, labelOrder{name: name, index: idx})
	}

	for i := 0; i < len(orderedLabels); i++ {
		for j := i + 1; j < len(orderedLabels); j++ {
			if orderedLabels[j].index < orderedLabels[i].index {
				orderedLabels[i], orderedLabels[j] = orderedLabels[j], orderedLabels[i]
			}
		}
	}

	valueIdx := 0
	for _, label := range orderedLabels {
		if valueIdx >= len(valueQueue) {
			break
		}

		switch label.name {
		case "VAT_INCLUSIVE":
			data.VATInclusive = valueQueue[valueIdx]
			valueIdx++
		case "LESS_VAT":
			data.LessVAT = valueQueue[valueIdx]
			valueIdx++
		case "LESS_WITHHOLDING":
			data.LessWithholding = valueQueue[valueIdx]
			valueIdx++
		case "NET_OF_VAT":
			data.AmountNetOfVAT = valueQueue[valueIdx]
			valueIdx++
		case "ADD_VAT":
			data.AddVAT = valueQueue[valueIdx]
			valueIdx++
		case "TOTAL_AMOUNT":
			data.Total = valueQueue[valueIdx]
			valueIdx++
		}
	}
}

func isNumericValue(s string) bool {
	if s == "" {
		return false
	}
	hasDigit := false
	for _, c := range s {
		if c >= '0' && c <= '9' {
			hasDigit = true
		} else if c != ',' && c != '.' && c != ' ' {
			return false
		}
	}
	return hasDigit
}

func findSectionBounds(lines []string, startMarker, endMarker string) (startIdx, endIdx int) {
	startIdx = -1
	endIdx = -1

	upperStartMarker := strings.ToUpper(startMarker)
	upperEndMarker := strings.ToUpper(endMarker)

	for i, line := range lines {
		upperLine := strings.ToUpper(strings.TrimSpace(line))
		if startIdx == -1 && strings.Contains(upperLine, upperStartMarker) {
			startIdx = i
		} else if startIdx != -1 && strings.Contains(upperLine, upperEndMarker) {
			endIdx = i
			break
		}
	}

	return startIdx, endIdx
}

func combineNumbers(lines []string, startIdx int) (combined string, skipCount int) {
	if startIdx >= len(lines) {
		return "", 0
	}

	currentLine := strings.TrimSpace(lines[startIdx])

	// Case 1: Current line has comma with incomplete digits (e.g., "6,2")
	// Next line should have remaining digits (e.g., "00")
	if strings.Contains(currentLine, ",") {
		commaIdx := strings.LastIndex(currentLine, ",")
		afterComma := currentLine[commaIdx+1:]

		// Check if we have incomplete digits after comma (1-2 digits means incomplete)
		if len(afterComma) >= 1 && len(afterComma) <= 2 {
			if startIdx+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[startIdx+1])
				// Check if next line starts with digits
				if len(nextLine) > 0 && nextLine[0] >= '0' && nextLine[0] <= '9' {
					beforeComma := currentLine[:commaIdx]
					digitsNeeded := 3 - len(afterComma)
					if digitsNeeded > 0 && digitsNeeded <= len(nextLine) {
						combined = beforeComma + afterComma + nextLine[:digitsNeeded]
						return combined, 1
					}
				}
			}
		}
		return currentLine, 0
	}

	// Case 2: Current line is all digits (e.g., "200")
	// Next line might have comma with incomplete digits (e.g., "6,2")
	// OCR split "6,200" as "200" then "6,2"
	if startIdx+1 < len(lines) {
		nextLine := strings.TrimSpace(lines[startIdx+1])
		if strings.Contains(nextLine, ",") {
			commaIdx := strings.LastIndex(nextLine, ",")
			afterComma := nextLine[commaIdx+1:]

			// Check if next line has incomplete digits after comma
			if len(afterComma) >= 1 && len(afterComma) <= 2 {
				beforeComma := nextLine[:commaIdx]
				digitsNeeded := 3 - len(afterComma)

				// Check if current line has enough digits to complete the number
				if digitsNeeded > 0 && digitsNeeded <= len(currentLine) {
					// Take needed digits from END of current line
					combined = beforeComma + afterComma + currentLine[len(currentLine)-digitsNeeded:]
					return combined, 1
				}
			}
		}
	}

	return currentLine, 0
}

func extractValueAfterLabel(text string, label string) string {
	upperText := strings.ToUpper(text)
	upperLabel := strings.ToUpper(label)

	patterns := []string{
		upperLabel + ":",
		upperLabel + " :",
		upperLabel,
	}

	for _, pattern := range patterns {
		if idx := strings.Index(upperText, pattern); idx != -1 {
			startIdx := idx + len(pattern)
			if startIdx < len(text) {
				value := strings.TrimSpace(text[startIdx:])
				if value != "" && value != ":" {
					return value
				}
			}
		}
	}

	return ""
}

func formatTIN(tin string) string {
	tin = strings.ReplaceAll(tin, " - ", "-")
	tin = strings.ReplaceAll(tin, " -", "-")
	tin = strings.ReplaceAll(tin, "- ", "-")
	tin = strings.ReplaceAll(tin, " ", "-")
	return tin
}

var _ scanner.IReceiptScanner = (*GoogleVisionScanner)(nil)

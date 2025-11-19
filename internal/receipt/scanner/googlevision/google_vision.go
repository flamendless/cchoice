package googlevision

import (
	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/receipt"
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

type WordInfo struct {
	Text       string
	Confidence float32
	X          float32
	Y          float32
}

func validate() {
	cfg := conf.Conf()
	if cfg.OCRService != receipt.RECEIPT_SCANNER_GOOGLEVISION.String() {
		panic(errs.ErrGVisionServiceInit)
	}
	if cfg.GoogleVisionConfig.APIKey == "" {
		panic(errs.ErrGVisionAPIKeyRequired)
	}
}

func MustInit() *GoogleVisionScanner {
	validate()
	cfg := conf.Conf()
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
	imageContext := &visionpb.ImageContext{
		LanguageHints: []string{"en", "tl"},
	}

	request := &visionpb.AnnotateImageRequest{
		Image: image,
		Features: []*visionpb.Feature{
			{
				Type: visionpb.Feature_DOCUMENT_TEXT_DETECTION,
			},
		},
		ImageContext: imageContext,
	}

	response, err := g.client.AnnotateImage(ctx, request)
	if err != nil {
		return nil, errors.Join(errs.ErrGVisionAPI, err)
	}

	if response.Error != nil {
		return nil, errors.Join(errs.ErrGVisionAPI, errors.New(response.Error.Message))
	}

	if response.FullTextAnnotation == nil || response.FullTextAnnotation.Text == "" {
		return nil, errs.ErrReceiptNoTextFound
	}

	fullText := response.FullTextAnnotation.Text
	logs.Log().Info("Text extracted from receipt", zap.Int("length", len(fullText)))

	structuredData := g.extractStructuredData(response.FullTextAnnotation)

	receiptData := g.parseReceiptWithStructuredData(fullText, structuredData)
	receiptData.RawText = fullText

	return receiptData, nil
}

type StructuredData struct {
	Words []WordInfo
	Rows  []Row
}

type Row struct {
	Y     float32
	Words []WordInfo
}

func (g *GoogleVisionScanner) extractStructuredData(fullAnnotation *visionpb.TextAnnotation) *StructuredData {
	if len(fullAnnotation.Pages) == 0 {
		return &StructuredData{}
	}

	var words []WordInfo
	minConfidence := float32(0.85)

	for _, page := range fullAnnotation.Pages {
		for _, block := range page.Blocks {
			for _, paragraph := range block.Paragraphs {
				for _, word := range paragraph.Words {
					vertices := word.BoundingBox.Vertices
					if len(vertices) < 4 {
						continue
					}

					avgX := (vertices[0].X + vertices[1].X + vertices[2].X + vertices[3].X) / 4
					avgY := (vertices[0].Y + vertices[1].Y + vertices[2].Y + vertices[3].Y) / 4

					var wordText string
					for _, symbol := range word.Symbols {
						wordText += symbol.Text
					}

					if word.Confidence >= minConfidence {
						words = append(words, WordInfo{
							Text:       wordText,
							Confidence: word.Confidence,
							X:          float32(avgX),
							Y:          float32(avgY),
						})
					}
				}
			}
		}
	}

	rows := g.organizeIntoRows(words)

	logs.Log().Info("Structured data extracted",
		zap.Int("total_words", len(words)),
		zap.Int("total_rows", len(rows)),
		zap.Float64("min_confidence", float64(minConfidence)),
		zap.Bool("has_data", len(words) > 0 && len(rows) > 0))

	return &StructuredData{
		Words: words,
		Rows:  rows,
	}
}

func (g *GoogleVisionScanner) organizeIntoRows(words []WordInfo) []Row {
	const rowThreshold = float32(20.0)

	var rows []Row
	for _, word := range words {
		placed := false
		for i := range rows {
			if abs(rows[i].Y-word.Y) < rowThreshold {
				rows[i].Words = append(rows[i].Words, word)
				placed = true
				break
			}
		}
		if !placed {
			rows = append(rows, Row{Y: word.Y, Words: []WordInfo{word}})
		}
	}

	for i := 0; i < len(rows); i++ {
		for j := i + 1; j < len(rows); j++ {
			if rows[j].Y < rows[i].Y {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
	}

	for i := range rows {
		for j := 0; j < len(rows[i].Words); j++ {
			for k := j + 1; k < len(rows[i].Words); k++ {
				if rows[i].Words[k].X < rows[i].Words[j].X {
					rows[i].Words[j], rows[i].Words[k] = rows[i].Words[k], rows[i].Words[j]
				}
			}
		}
	}

	logs.Log().Info("Table structure analyzed", zap.Int("rows_detected", len(rows)))
	return rows
}

func (g *GoogleVisionScanner) parseReceiptWithStructuredData(text string, structuredData *StructuredData) *scanner.ReceiptData {
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

	if len(structuredData.Rows) > 0 {
		logs.Log().Info("Using structured data parsing",
			zap.Int("available_rows", len(structuredData.Rows)),
			zap.Int("available_words", len(structuredData.Words)),
		)
		parseLineItemsWithStructuredData(lines, data, structuredData)
	} else {
		logs.Log().Info("Falling back to text-only parsing (no structured data available)")
		parseLineItems(lines, data)
	}

	parseTotalSales(lines, data)

	return data
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func parseMerchant(lines []string, data *scanner.ReceiptData) {
	startIdx, endIdx := findSectionBounds(lines, "CHOICE", "SALES")
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
		case strings.Contains(upperLine, "CHOICE") && (strings.Contains(upperLine, "CONSTRUCTION") || strings.Contains(upperLine, "C-CHOICE")):
			data.MerchantName = upperLine
		case strings.Contains(upperLine, "BLK") || strings.Contains(upperLine, "LT") || strings.Contains(upperLine, "PHILIPPINES") || strings.Contains(upperLine, "PHILIPPINE") || strings.Contains(upperLine, "CAVITE"):
			data.MerchantAddress = upperLine
		case strings.Contains(upperLine, "VAT") || isTINLine(upperLine):
			if tin := extractValueAfterLabel(upperLine, "TIN"); tin != "" {
				data.MerchantTIN = formatTIN(tin)
			}
		case strings.Contains(upperLine, "PROP") || strings.Contains(upperLine, "-PROP"):
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
			if val := extractValueAfterLabel(upperLine, "SOLD TO"); val != "" {
				data.SoldTo = val
			} else if i+1 < len(lines) {
				data.SoldTo = strings.ToUpper(strings.TrimSpace(lines[i+1]))
			}
		case isTINLine(upperLine):
			if tin := extractValueAfterLabel(upperLine, "TIN"); tin != "" {
				data.CustomerTIN = formatTIN(tin)
			}
		case strings.Contains(upperLine, "ADDRESS"):
			preaddress := extractValueAfterLabel(upperLine, "ADDRESS")
			if preaddress != "" {
				data.CustomerAddress = strings.ToUpper(preaddress)
			} else if i+1 < len(lines) {
				data.CustomerAddress = strings.ToUpper(strings.TrimSpace(lines[i+1]))
			}
			data.CustomerAddress = strings.ReplaceAll(data.CustomerAddress, "SUBOY.", "SUBD.")
			data.CustomerAddress = strings.ReplaceAll(data.CustomerAddress, "BROY", "BRGY.")
		case strings.Contains(upperLine, "DATE") && !strings.Contains(upperLine, "RATED"):
			if date := extractValueAfterLabel(upperLine, "DATE"); date != "" {
				data.Date = strings.ToUpper(date)
			}
		}
	}
}

func parseLineItemsWithStructuredData(lines []string, data *scanner.ReceiptData, structuredData *StructuredData) {
	startIdx := -1
	for i, line := range lines {
		upperLine := strings.ToUpper(strings.TrimSpace(line))
		if strings.Contains(upperLine, "ITEM DESCRIPTION") || strings.Contains(upperLine, "NATURE OF SERVICE") {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		logs.Log().Info("Table section not found, using fallback text parsing")
		parseLineItems(lines, data)
		return
	}

	endIdx := -1
	for i := startIdx + 1; i < len(lines); i++ {
		upperLine := strings.ToUpper(strings.TrimSpace(lines[i]))
		if (strings.Contains(upperLine, "TOTAL") && strings.Contains(upperLine, "SALE")) ||
			(strings.Contains(upperLine, "LESS") && strings.Contains(upperLine, "VAT")) {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		endIdx = len(lines)
	}

	logs.Log().Info("Using structured data for line item parsing",
		zap.Int("start_line", startIdx),
		zap.Int("end_line", endIdx))

	var tableRows []Row
	for _, row := range structuredData.Rows {
		rowText := ""
		for _, word := range row.Words {
			rowText += strings.ToUpper(word.Text) + " "
		}

		for i := startIdx; i <= endIdx && i < len(lines); i++ {
			if strings.Contains(rowText, strings.ToUpper(strings.TrimSpace(lines[i]))) {
				tableRows = append(tableRows, row)
				break
			}
		}
	}

	logs.Log().Info("Table rows identified", zap.Int("table_rows", len(tableRows)))

	if len(tableRows) > 0 {
		logs.Log().Info("Structured table data is available but using text parsing for now",
			zap.String("note", "Future enhancement: use X-coordinates for column detection"))
	}

	parseLineItems(lines, data)

	logs.Log().Info("Line items extracted",
		zap.Int("item_count", len(data.Items)),
		zap.Bool("used_structured_data", len(tableRows) > 0))
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
		if (strings.Contains(upperLine, "TOTAL") && strings.Contains(upperLine, "SALE")) ||
			(strings.Contains(upperLine, "LESS") && strings.Contains(upperLine, "VAT")) {
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
	priceIdx := -1
	amountIdx := -1

	// Find all header positions
	for i := startIdx; i < endIdx; i++ {
		upperLine := strings.ToUpper(strings.TrimSpace(lines[i]))
		switch {
		case upperLine == "QUANTITY":
			quantityIdx = i
		case strings.Contains(upperLine, "UNIT") && strings.Contains(upperLine, "COST"):
			unitCostIdx = i
		case upperLine == "PRICE":
			priceIdx = i
		case upperLine == "AMOUNT" && !strings.Contains(upperLine, "TOTAL"):
			amountIdx = i
		}
	}

	logs.Log().Info("Header positions found",
		zap.Int("item_desc", itemDescIdx),
		zap.Int("quantity", quantityIdx),
		zap.Int("unit_cost", unitCostIdx),
		zap.Int("price", priceIdx),
		zap.Int("amount", amountIdx))

	// Find where ALL headers end (max of all header positions)
	headerEndIdx := max(amountIdx, max(priceIdx, max(unitCostIdx, max(quantityIdx, itemDescIdx))))

	logs.Log().Info("Headers end at line", zap.Int("header_end", headerEndIdx))

	// Extract item names starting AFTER all headers end
	// Items are in a columnar layout: all data starts after the last header line
	itemsStartIdx := headerEndIdx + 1

	for i := itemsStartIdx; i < endIdx; i++ {
		trimmedLine := strings.TrimSpace(lines[i])
		if trimmedLine == "" {
			continue
		}
		upperLine := strings.ToUpper(trimmedLine)

		// Stop at totals section
		if (strings.Contains(upperLine, "TOTAL") && strings.Contains(upperLine, "SALE")) ||
			strings.Contains(upperLine, "VAT INCLUSIVE") ||
			(strings.Contains(upperLine, "LESS") && strings.Contains(upperLine, "VAT")) {
			break
		}

		// Skip if it's numeric-only (likely quantity/price/amount, not an item name)
		if isNumericValue(trimmedLine) {
			continue
		}

		// Skip if it's a header line
		if strings.Contains(upperLine, "QUANTITY") || strings.Contains(upperLine, "UNIT") ||
			strings.Contains(upperLine, "PRICE") || strings.Contains(upperLine, "AMOUNT") ||
			strings.Contains(upperLine, "COST") {
			continue
		}

		// Skip very short strings (likely OCR errors or single letters)
		if len(trimmedLine) < 3 {
			continue
		}

		// Skip if it looks like a price/amount (numbers with dots/commas and dashes)
		digitCount := 0
		for _, c := range trimmedLine {
			if c >= '0' && c <= '9' {
				digitCount++
			}
		}
		// If more than 50% digits, it's probably a number
		if float64(digitCount)/float64(len(trimmedLine)) > 0.5 {
			continue
		}

		// This should be an item name
		lineItem := scanner.LineItem{Name: upperLine}
		data.Items = append(data.Items, lineItem)
	}

	logs.Log().Info("Items extracted by name", zap.Int("count", len(data.Items)))

	// Extract quantities
	if quantityIdx != -1 {
		nextHeaderIdx := endIdx
		switch {
		case unitCostIdx != -1:
			nextHeaderIdx = unitCostIdx
		case priceIdx != -1:
			nextHeaderIdx = priceIdx
		case amountIdx != -1:
			nextHeaderIdx = amountIdx
		}

		lineItemIdx := 0
		for i := quantityIdx + 1; i < nextHeaderIdx && lineItemIdx < len(data.Items); i++ {
			trimmedLine := strings.TrimSpace(lines[i])
			if trimmedLine == "" {
				continue
			}
			upperLine := strings.ToUpper(trimmedLine)
			// Skip header lines
			if strings.Contains(upperLine, "UNIT") || strings.Contains(upperLine, "COST") ||
				strings.Contains(upperLine, "PRICE") || strings.Contains(upperLine, "AMOUNT") {
				continue
			}
			data.Items[lineItemIdx].Quantity = trimmedLine
			lineItemIdx++
		}
		// Fill remaining items with default quantity
		for lineItemIdx < len(data.Items) {
			data.Items[lineItemIdx].Quantity = "1"
			lineItemIdx++
		}
	}

	// Extract prices (from Unit Cost or Price section)
	priceStartIdx := -1
	priceEndIdx := endIdx
	if unitCostIdx != -1 {
		priceStartIdx = unitCostIdx
		if priceIdx != -1 {
			priceEndIdx = priceIdx
		} else if amountIdx != -1 {
			priceEndIdx = amountIdx
		}
	} else if priceIdx != -1 {
		priceStartIdx = priceIdx
		if amountIdx != -1 {
			priceEndIdx = amountIdx
		}
	}

	if priceStartIdx != -1 {
		lineItemIdx := 0
		i := priceStartIdx + 1
		for i < priceEndIdx && lineItemIdx < len(data.Items) {
			trimmedLine := strings.TrimSpace(lines[i])
			if trimmedLine == "" {
				i++
				continue
			}
			upperLine := strings.ToUpper(trimmedLine)
			// Skip header lines
			if strings.Contains(upperLine, "PRICE") || strings.Contains(upperLine, "COST") ||
				strings.Contains(upperLine, "AMOUNT") {
				i++
				continue
			}
			// Check if it's a numeric value
			if isNumericValue(trimmedLine) {
				combined, skipCount := combineNumbers(lines, i)
				data.Items[lineItemIdx].Price = combined
				lineItemIdx++
				i += 1 + skipCount
			} else {
				i++
			}
		}
	}

	// Extract amounts/subtotals
	if amountIdx != -1 {
		lineItemIdx := 0
		i := amountIdx + 1
		for i < endIdx && lineItemIdx < len(data.Items) {
			trimmedLine := strings.TrimSpace(lines[i])
			if trimmedLine == "" {
				i++
				continue
			}
			upperLine := strings.ToUpper(trimmedLine)
			// Stop if we hit the totals section
			if (strings.Contains(upperLine, "TOTAL") && strings.Contains(upperLine, "SALE")) ||
				strings.Contains(upperLine, "VAT INCLUSIVE") ||
				(strings.Contains(upperLine, "LESS") && strings.Contains(upperLine, "VAT")) {
				break
			}
			// Check if it's a numeric value
			if isNumericValue(trimmedLine) {
				combined, skipCount := combineNumbers(lines, i)
				data.Items[lineItemIdx].Subtotal = combined
				lineItemIdx++
				i += 1 + skipCount
			} else {
				i++
			}
		}
	}
}

func parseTotalSales(lines []string, data *scanner.ReceiptData) {
	startIdx := -1
	for i, line := range lines {
		upperLine := strings.ToUpper(strings.TrimSpace(line))
		if strings.Contains(upperLine, "TOTAL SALES") || strings.Contains(upperLine, "VAT INCLUSIVE") {
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
		case strings.Contains(upperLine, "TOTAL SALES") || strings.Contains(upperLine, "VAT INCLUSIVE"):
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

func isTINLine(line string) bool {
	upperLine := strings.ToUpper(line)
	if !strings.Contains(upperLine, "TIN") {
		return false
	}

	patterns := []string{
		"TIN:",
		"TIN :",
		"TIN NO",
		"TIN#",
		" TIN ",
		" TIN:",
	}

	for _, pattern := range patterns {
		if strings.Contains(upperLine, pattern) {
			return true
		}
	}

	if strings.HasPrefix(upperLine, "TIN") && len(upperLine) > 3 {
		nextChar := upperLine[3]
		if nextChar == ':' || nextChar == ' ' || nextChar == '#' || (nextChar >= '0' && nextChar <= '9') {
			return true
		}
	}

	return false
}

var _ scanner.IReceiptScanner = (*GoogleVisionScanner)(nil)

package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

// RenderInvoicePDF renders an invoice (with its config and lines) into PDF bytes.
// Amounts are rendered with the currency code prefix (e.g. "PHP 1,500.00")
// because the core PDF fonts cannot render currency glyphs such as the peso sign.
func RenderInvoicePDF(config InvoiceConfig, invoice Invoice, lines []InvoiceLine) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	const pageWidth = 180.0 // 210 - 2*15 margins
	currency := config.Currency
	if strings.TrimSpace(currency) == "" {
		currency = invoice.Currency
	}

	// Header: logo (left) + INVOICE title (right)
	startY := pdf.GetY()
	logoRendered := drawInvoiceLogo(pdf, config, 15, startY)

	pdf.SetXY(15, startY)
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(40, 40, 40)
	pdf.CellFormat(pageWidth, 12, "INVOICE", "", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(90, 90, 90)
	pdf.CellFormat(pageWidth, 6, tr(invoice.InvoiceNumber), "", 1, "R", false, 0, "")

	if logoRendered {
		if y := startY + 26; pdf.GetY() < y {
			pdf.SetY(y)
		}
	}
	pdf.Ln(4)

	// Business (seller) block
	pdf.SetTextColor(40, 40, 40)
	pdf.SetFont("Arial", "B", 12)
	businessName := config.BusinessName
	if strings.TrimSpace(businessName) == "" {
		businessName = "C-Choice Construction Supply Shop"
	}
	pdf.CellFormat(pageWidth, 6, tr(businessName), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(90, 90, 90)
	for _, l := range businessLines(config) {
		pdf.MultiCell(pageWidth, 5, tr(l), "", "L", false)
	}
	pdf.Ln(3)

	// Bill To + Invoice meta side by side
	metaY := pdf.GetY()
	pdf.SetTextColor(120, 120, 120)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(pageWidth/2, 5, "BILL TO", "", 0, "L", false, 0, "")
	pdf.CellFormat(pageWidth/2, 5, "DETAILS", "", 1, "L", false, 0, "")

	pdf.SetTextColor(40, 40, 40)
	pdf.SetFont("Arial", "B", 10)
	recipientName := invoice.RecipientName
	if strings.TrimSpace(recipientName) == "" {
		recipientName = "-"
	}
	pdf.SetXY(15, pdf.GetY())
	pdf.CellFormat(pageWidth/2, 5, tr(recipientName), "", 0, "L", false, 0, "")

	// Right column meta rows
	metaRows := [][2]string{
		{"Invoice No:", invoice.InvoiceNumber},
		{"Issue Date:", invoice.IssueDate},
		{"Due Date:", orDash(invoice.DueDate)},
		{"Status:", invoice.Status.String()},
	}
	rightX := 15 + pageWidth/2
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(90, 90, 90)
	pdf.SetXY(rightX, metaY+5)
	for _, mr := range metaRows {
		pdf.SetX(rightX)
		pdf.CellFormat(30, 5, mr[0], "", 0, "L", false, 0, "")
		pdf.CellFormat(pageWidth/2-30, 5, tr(mr[1]), "", 1, "R", false, 0, "")
	}
	metaBottom := pdf.GetY()

	// Left column recipient details
	pdf.SetXY(15, metaY+10)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(90, 90, 90)
	for _, l := range recipientLines(invoice) {
		pdf.SetX(15)
		pdf.MultiCell(pageWidth/2-5, 5, tr(l), "", "L", false)
	}
	if pdf.GetY() < metaBottom {
		pdf.SetY(metaBottom)
	}
	pdf.Ln(4)

	// Line items table
	drawLineItemsTable(pdf, tr, lines, currency, pageWidth)

	// Totals (right aligned)
	pdf.Ln(2)
	totalsX := 15 + pageWidth - 70
	drawTotalRow(pdf, tr, totalsX, "Subtotal", formatMoneyPlain(invoice.SubtotalRaw, currency), false)
	vatLabel := "VAT"
	if strings.TrimSpace(invoice.VATPercentage) != "" && invoice.VATPercentage != "0" {
		vatLabel = fmt.Sprintf("VAT (%s%%)", invoice.VATPercentage)
	}
	drawTotalRow(pdf, tr, totalsX, vatLabel, formatMoneyPlain(invoice.VATAmountRaw, currency), false)
	drawTotalRow(pdf, tr, totalsX, "Total", formatMoneyPlain(invoice.TotalRaw, currency), true)

	// Notes + footer
	if strings.TrimSpace(invoice.Notes) != "" {
		pdf.Ln(6)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(pageWidth, 5, "NOTES", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(90, 90, 90)
		pdf.MultiCell(pageWidth, 5, tr(invoice.Notes), "", "L", false)
	}
	if strings.TrimSpace(config.FooterNotes) != "" {
		pdf.Ln(4)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(140, 140, 140)
		pdf.MultiCell(pageWidth, 4, tr(config.FooterNotes), "", "C", false)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawInvoiceLogo(pdf *fpdf.Fpdf, config InvoiceConfig, x, y float64) bool {
	const maxW = 45.0
	const maxH = 22.0

	if path := strings.TrimSpace(config.LogoPath); path != "" {
		if _, err := os.Stat(path); err == nil {
			imgType := imageTypeFromExt(path)
			if imgType != "" {
				pdf.ImageOptions(path, x, y, maxW, 0, false, fpdf.ImageOptions{ImageType: imgType, ReadDpi: true}, 0, "")
				return true
			}
		}
	}

	if url := strings.TrimSpace(config.LogoURL); strings.HasPrefix(url, "http") {
		data, imgType := fetchImage(url)
		if len(data) > 0 && imgType != "" {
			name := "invoice_logo_remote"
			pdf.RegisterImageOptionsReader(name, fpdf.ImageOptions{ImageType: imgType, ReadDpi: true}, bytes.NewReader(data))
			pdf.ImageOptions(name, x, y, maxW, 0, false, fpdf.ImageOptions{ImageType: imgType, ReadDpi: true}, 0, "")
			return true
		}
	}

	_ = maxH
	return false
}

func fetchImage(url string) ([]byte, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, ""
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ""
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, ""
	}
	imgType := imageTypeFromContentType(resp.Header.Get("Content-Type"))
	if imgType == "" {
		imgType = imageTypeFromExt(url)
	}
	return data, imgType
}

func imageTypeFromExt(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		return "PNG"
	case ".jpg", ".jpeg":
		return "JPG"
	default:
		return ""
	}
}

func imageTypeFromContentType(ct string) string {
	switch {
	case strings.Contains(ct, "png"):
		return "PNG"
	case strings.Contains(ct, "jpeg"), strings.Contains(ct, "jpg"):
		return "JPG"
	default:
		return ""
	}
}

func drawLineItemsTable(pdf *fpdf.Fpdf, tr func(string) string, lines []InvoiceLine, currency string, pageWidth float64) {
	descW := pageWidth * 0.46
	qtyW := pageWidth * 0.12
	priceW := pageWidth * 0.21
	amtW := pageWidth * 0.21

	pdf.SetFillColor(247, 239, 234)
	pdf.SetTextColor(70, 70, 70)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(descW, 8, "Description", "", 0, "L", true, 0, "")
	pdf.CellFormat(qtyW, 8, "Qty", "", 0, "C", true, 0, "")
	pdf.CellFormat(priceW, 8, "Unit Price", "", 0, "R", true, 0, "")
	pdf.CellFormat(amtW, 8, "Amount", "", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(50, 50, 50)
	for _, l := range lines {
		x := pdf.GetX()
		y := pdf.GetY()
		pdf.MultiCell(descW, 6, tr(l.Description), "B", "L", false)
		rowH := pdf.GetY() - y
		if rowH < 6 {
			rowH = 6
		}
		pdf.SetXY(x+descW, y)
		pdf.CellFormat(qtyW, rowH, fmt.Sprintf("%d", l.Quantity), "B", 0, "C", false, 0, "")
		pdf.CellFormat(priceW, rowH, formatMoneyPlain(l.UnitPriceRaw, currency), "B", 0, "R", false, 0, "")
		pdf.CellFormat(amtW, rowH, formatMoneyPlain(l.LineTotalRaw, currency), "B", 1, "R", false, 0, "")
	}
}

func drawTotalRow(pdf *fpdf.Fpdf, tr func(string) string, x float64, label, value string, bold bool) {
	pdf.SetX(x)
	if bold {
		pdf.SetFont("Arial", "B", 11)
		pdf.SetTextColor(246, 116, 47)
	} else {
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(90, 90, 90)
	}
	pdf.CellFormat(35, 7, tr(label), "", 0, "L", false, 0, "")
	pdf.CellFormat(35, 7, value, "", 1, "R", false, 0, "")
}

func businessLines(config InvoiceConfig) []string {
	lines := make([]string, 0, 6)
	if v := strings.TrimSpace(config.Address); v != "" {
		lines = append(lines, v)
	}
	if v := strings.TrimSpace(config.TIN); v != "" {
		lines = append(lines, "TIN: "+v)
	}
	if v := strings.TrimSpace(config.VATRegistration); v != "" {
		lines = append(lines, "VAT Reg: "+v)
	}
	contact := make([]string, 0, 2)
	if v := strings.TrimSpace(config.ContactNumber); v != "" {
		contact = append(contact, v)
	}
	if v := strings.TrimSpace(config.Email); v != "" {
		contact = append(contact, v)
	}
	if len(contact) > 0 {
		lines = append(lines, strings.Join(contact, "  |  "))
	}
	if v := strings.TrimSpace(config.Website); v != "" {
		lines = append(lines, v)
	}
	return lines
}

func recipientLines(invoice Invoice) []string {
	lines := make([]string, 0, 5)
	if v := strings.TrimSpace(invoice.RecipientAddress); v != "" {
		lines = append(lines, v)
	}
	if v := strings.TrimSpace(invoice.RecipientContactNumber); v != "" {
		lines = append(lines, v)
	}
	if v := strings.TrimSpace(invoice.RecipientEmail); v != "" {
		lines = append(lines, v)
	}
	if v := strings.TrimSpace(invoice.RecipientTIN); v != "" {
		lines = append(lines, "TIN: "+v)
	}
	return lines
}

func orDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}

// formatMoneyPlain formats a centavo amount as "CUR 1,234.56" without any
// currency glyph so it renders correctly with the core PDF fonts.
func formatMoneyPlain(centavos int64, currency string) string {
	negative := centavos < 0
	if negative {
		centavos = -centavos
	}
	pesos := centavos / 100
	frac := centavos % 100

	intStr := fmt.Sprintf("%d", pesos)
	var grouped strings.Builder
	n := len(intStr)
	for i, ch := range intStr {
		if i > 0 && (n-i)%3 == 0 {
			grouped.WriteByte(',')
		}
		grouped.WriteRune(ch)
	}

	sign := ""
	if negative {
		sign = "-"
	}
	cur := strings.TrimSpace(currency)
	if cur != "" {
		cur += " "
	}
	return fmt.Sprintf("%s%s%s.%02d", cur, sign, grouped.String(), frac)
}

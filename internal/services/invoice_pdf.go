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

	"cchoice/internal/constants"
	"cchoice/internal/enums"

	"github.com/go-pdf/fpdf"
)

func RenderInvoicePDF(config InvoiceConfig, invoice Invoice, lines []InvoiceLine) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 12)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	const pageWidth = 186.0
	currency := config.Currency
	if strings.TrimSpace(currency) == "" {
		currency = invoice.Currency
	}

	startY := pdf.GetY()
	logoRendered := drawInvoiceLogo(pdf, config, 12, startY)

	pdf.SetXY(12, startY)
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(20, 20, 20)
	pdf.CellFormat(pageWidth, 8, tr("SALES INVOICE"), "", 1, "C", false, 0, "")

	if logoRendered {
		if y := startY + 24; pdf.GetY() < y {
			pdf.SetY(y)
		}
	}
	pdf.Ln(2)

	drawSellerHeader(pdf, tr, config, pageWidth)
	drawTransactionType(pdf, tr, invoice.TransactionType, pageWidth)
	drawCustomerBlock(pdf, tr, invoice, pageWidth)
	drawBIRLineItemsTable(pdf, tr, lines, currency, pageWidth)
	drawBIRSummary(pdf, tr, invoice, currency, pageWidth)
	drawBIRFooter(pdf, tr, config, invoice, pageWidth)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawSellerHeader(pdf *fpdf.Fpdf, tr func(string) string, config InvoiceConfig, pageWidth float64) {
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(30, 30, 30)
	businessName := config.BusinessName
	if strings.TrimSpace(businessName) == "" {
		businessName = "C-Choice Construction Supply Shop"
	}
	pdf.CellFormat(pageWidth, 5, tr(businessName), "", 1, "C", false, 0, "")

	if v := strings.TrimSpace(config.ProprietorName); v != "" {
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(60, 60, 60)
		pdf.CellFormat(pageWidth, 4, tr("Proprietor: "+v), "", 1, "C", false, 0, "")
	}

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(80, 80, 80)
	for _, l := range businessLines(config) {
		pdf.CellFormat(pageWidth, 4, tr(l), "", 1, "C", false, 0, "")
	}
	pdf.Ln(2)
}

func drawTransactionType(pdf *fpdf.Fpdf, tr func(string) string, txType enums.InvoiceTransactionType, pageWidth float64) {
	isCash := txType == enums.INVOICE_TRANSACTION_CASH_SALES
	isCharge := txType == enums.INVOICE_TRANSACTION_CHARGE_SALES
	cashMark := "[ ]"
	chargeMark := "[ ]"
	if isCash {
		cashMark = "[X]"
	}
	if isCharge {
		chargeMark = "[X]"
	}
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(pageWidth, 5, tr(fmt.Sprintf("%s Cash Sales     %s Charge Sales", cashMark, chargeMark)), "", 1, "L", false, 0, "")
	pdf.Ln(1)
}

func drawCustomerBlock(pdf *fpdf.Fpdf, tr func(string) string, invoice Invoice, pageWidth float64) {
	leftW := pageWidth * 0.62
	rightW := pageWidth - leftW
	y := pdf.GetY()

	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(leftW, 4, tr("SOLD TO:"), "", 0, "L", false, 0, "")
	pdf.CellFormat(rightW, 4, tr("INVOICE NO."), "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(30, 30, 30)
	regName := invoice.RecipientRegisteredName
	if strings.TrimSpace(regName) == "" {
		regName = invoice.RecipientName
	}
	pdf.CellFormat(leftW, 5, tr(regName), "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(rightW, 5, tr(invoice.InvoiceNumber), "", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(70, 70, 70)
	for _, l := range recipientLines(invoice) {
		pdf.CellFormat(leftW, 4, tr(l), "", 0, "L", false, 0, "")
		pdf.CellFormat(rightW, 4, "", "", 1, "L", false, 0, "")
	}

	metaY := y
	pdf.SetXY(12+leftW, metaY+9)
	meta := [][2]string{
		{"Date:", invoice.IssueDate},
		{"Due:", orDash(invoice.DueDate)},
		{"TIN:", orDash(invoice.RecipientTIN)},
	}
	for _, mr := range meta {
		pdf.SetX(12 + leftW)
		pdf.CellFormat(18, 4, tr(mr[0]), "", 0, "L", false, 0, "")
		pdf.CellFormat(rightW-18, 4, tr(mr[1]), "", 1, "R", false, 0, "")
	}
	if pdf.GetY() < y+22 {
		pdf.SetY(y + 22)
	}
	pdf.Ln(2)
}

func drawBIRLineItemsTable(pdf *fpdf.Fpdf, tr func(string) string, lines []InvoiceLine, currency string, pageWidth float64) {
	qtyW := pageWidth * 0.10
	unitW := pageWidth * 0.14
	descW := pageWidth * 0.46
	amtW := pageWidth * 0.15
	taxW := pageWidth - qtyW - unitW - descW - amtW

	pdf.SetFillColor(240, 240, 240)
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetFont("Arial", "B", 7)
	pdf.SetTextColor(50, 50, 50)
	headers := []struct {
		label string
		w     float64
		align string
	}{
		{"Qty", qtyW, "C"},
		{"Unit", unitW, "C"},
		{"Description", descW, "L"},
		{"Amount", amtW, "R"},
		{"Tax", taxW, "C"},
	}
	for _, h := range headers {
		pdf.CellFormat(h.w, 6, tr(h.label), "1", 0, h.align, true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(30, 30, 30)
	for _, l := range lines {
		x := pdf.GetX()
		y := pdf.GetY()
		pdf.MultiCell(descW, 5, tr(l.Description), "1", "L", false)
		rowH := pdf.GetY() - y
		if rowH < 5 {
			rowH = 5
		}
		pdf.SetXY(x, y)
		pdf.CellFormat(qtyW, rowH, fmt.Sprintf("%d", l.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(unitW, rowH, formatMoneyPlain(l.UnitPriceRaw, currency), "1", 0, "R", false, 0, "")
		pdf.SetXY(x+qtyW+unitW, y)
		pdf.CellFormat(descW, rowH, "", "1", 0, "L", false, 0, "")
		pdf.CellFormat(amtW, rowH, formatMoneyPlain(l.LineTotalRaw, currency), "1", 0, "R", false, 0, "")
		taxLabel := l.TaxType.Label()
		pdf.CellFormat(taxW, rowH, tr(taxLabel), "1", 1, "C", false, 0, "")
	}
	pdf.Ln(2)
}

func drawBIRSummary(pdf *fpdf.Fpdf, tr func(string) string, invoice Invoice, currency string, pageWidth float64) {
	leftW := pageWidth * 0.50
	rightW := pageWidth - leftW
	y := pdf.GetY()

	leftRows := [][2]string{
		{"VATable Sales", formatMoneyPlain(invoice.VATableSalesRaw, currency)},
		{"VAT-Exempt Sales", formatMoneyPlain(invoice.VATExemptSalesRaw, currency)},
		{"Zero Rated Sales", formatMoneyPlain(invoice.ZeroRatedSalesRaw, currency)},
		{"VAT Amount", formatMoneyPlain(invoice.VATAmountRaw, currency)},
		{"Total Sales", formatMoneyPlain(invoice.TotalSalesRaw, currency)},
	}
	rightRows := [][2]string{
		{"Total Sales (VAT Incl.)", formatMoneyPlain(invoice.TotalSalesVATInclusiveRaw, currency)},
		{"Less: VAT", formatMoneyPlain(invoice.LessVATRaw, currency)},
		{"Less: Withholding Tax", formatMoneyPlain(invoice.WithholdingTaxRaw, currency)},
		{"Amount Net of VAT", formatMoneyPlain(invoice.AmountNetOfVATRaw, currency)},
		{"Less: SC/PWD Discount", formatMoneyPlain(invoice.SCPWDDiscountRaw, currency)},
		{"Add: VAT", formatMoneyPlain(invoice.AddVATRaw, currency)},
		{"TOTAL AMOUNT DUE", formatMoneyPlain(invoice.TotalSalesVATInclusiveRaw, currency)},
	}

	drawSummaryColumn(pdf, tr, 12, y, leftW, leftRows, false)
	drawSummaryColumn(pdf, tr, 12+leftW, y, rightW, rightRows, true)

	bottom := pdf.GetY()
	if bottom < y+42 {
		pdf.SetY(y + 42)
	}
	pdf.Ln(2)
}

func drawSummaryColumn(pdf *fpdf.Fpdf, tr func(string) string, x, y, w float64, rows [][2]string, highlightLast bool) {
	pdf.SetXY(x, y)
	for i, row := range rows {
		isLast := i == len(rows)-1
		if highlightLast && isLast {
			pdf.SetFont("Arial", "B", 9)
			pdf.SetTextColor(200, 80, 20)
		} else {
			pdf.SetFont("Arial", "", 8)
			pdf.SetTextColor(60, 60, 60)
		}
		pdf.SetX(x)
		labelW := w * 0.58
		valW := w - labelW
		pdf.CellFormat(labelW, 5, tr(row[0]), "0", 0, "L", false, 0, "")
		pdf.CellFormat(valW, 5, row[1], "0", 1, "R", false, 0, "")
	}
	if pdf.GetY() > y {
		return
	}
	pdf.SetY(y + float64(len(rows))*5)
}

func drawBIRFooter(pdf *fpdf.Fpdf, tr func(string) string, config InvoiceConfig, invoice Invoice, pageWidth float64) {
	if v := strings.TrimSpace(invoice.ReceivedAmount); v != "" {
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(60, 60, 60)
		pdf.CellFormat(pageWidth, 4, tr("Received Amount: "+v), "", 1, "L", false, 0, "")
	}

	pdf.Ln(2)
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(80, 80, 80)
	footerParts := make([]string, 0, 4)
	if v := strings.TrimSpace(config.BIRAuthorityToPrint); v != "" {
		footerParts = append(footerParts, "BIR Authority to Print No. "+v)
	}
	if v := strings.TrimSpace(config.BIRDateIssued); v != "" {
		footerParts = append(footerParts, "Date Issued: "+v)
	}
	if v := strings.TrimSpace(config.BIRBookletsInfo); v != "" {
		footerParts = append(footerParts, v)
	}
	if len(footerParts) > 0 {
		pdf.MultiCell(pageWidth, 3.5, tr(strings.Join(footerParts, "  |  ")), "", "C", false)
	}

	if v := strings.TrimSpace(invoice.SCPWDIDNo); v != "" {
		pdf.Ln(1)
		pdf.CellFormat(pageWidth, 4, tr("SC/PWD ID No.: "+v), "", 1, "L", false, 0, "")
	}

	if strings.TrimSpace(invoice.Notes) != "" {
		pdf.Ln(2)
		pdf.SetFont("Arial", "B", 8)
		pdf.SetTextColor(100, 100, 100)
		pdf.CellFormat(pageWidth, 4, tr("Remarks"), "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8)
		pdf.MultiCell(pageWidth, 4, tr(invoice.Notes), "", "L", false)
	}

	if strings.TrimSpace(config.FooterNotes) != "" {
		pdf.Ln(2)
		pdf.SetFont("Arial", "I", 7)
		pdf.SetTextColor(120, 120, 120)
		pdf.MultiCell(pageWidth, 3.5, tr(config.FooterNotes), "", "C", false)
	}

	sigY := pdf.GetY() + 8
	if sigY > 250 {
		pdf.AddPage()
		sigY = pdf.GetY() + 8
	}
	colW := pageWidth / 3
	pdf.SetY(sigY)
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(80, 80, 80)
	for i, label := range []string{"Prepared by", "Checked by", "Received by"} {
		x := 12 + float64(i)*colW
		pdf.SetXY(x, sigY+12)
		pdf.CellFormat(colW-4, 4, tr(label), "T", 0, "C", false, 0, "")
	}
	pdf.SetY(sigY + 20)
}

func drawInvoiceLogo(pdf *fpdf.Fpdf, config InvoiceConfig, x, y float64) bool {
	const maxW = 40.0

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

	if data, imgType := fetchImage(constants.PathEmailLogoCDN); len(data) > 0 && imgType != "" {
		name := "invoice_logo_default"
		pdf.RegisterImageOptionsReader(name, fpdf.ImageOptions{ImageType: imgType, ReadDpi: true}, bytes.NewReader(data))
		pdf.ImageOptions(name, x, y, maxW, 0, false, fpdf.ImageOptions{ImageType: imgType, ReadDpi: true}, 0, "")
		return true
	}

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
	return lines
}

func orDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}

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

package services

import (
	"bytes"
	"strings"

	"github.com/go-pdf/fpdf"
)

const deliveryReceiptTableRows = 20

func RenderDeliveryReceiptPDF(config InvoiceConfig, receipt DeliveryReceipt, lines []DeliveryReceiptLine) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 12)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	const pageWidth = 186.0

	startY := pdf.GetY()
	logoRendered := drawInvoiceLogo(pdf, config, 12, startY)

	pdf.SetXY(12, startY)
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(20, 20, 20)
	pdf.CellFormat(pageWidth, 8, tr("DELIVERY RECEIPT"), "", 1, "C", false, 0, "")

	if logoRendered {
		if y := startY + 24; pdf.GetY() < y {
			pdf.SetY(y)
		}
	}
	pdf.Ln(2)

	drawSellerHeader(pdf, tr, config, pageWidth)
	drawDeliveryReceiptCustomerBlock(pdf, tr, receipt, pageWidth)
	drawDeliveryReceiptLineItemsTable(pdf, tr, lines, pageWidth)
	drawDeliveryReceiptFooter(pdf, tr, config, pageWidth)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawDeliveryReceiptCustomerBlock(pdf *fpdf.Fpdf, tr func(string) string, receipt DeliveryReceipt, pageWidth float64) {
	leftW := pageWidth * 0.62
	rightW := pageWidth - leftW
	y := pdf.GetY()

	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(leftW, 4, tr("DELIVERED TO:"), "", 0, "L", false, 0, "")
	pdf.CellFormat(rightW, 4, tr("RECEIPT NO."), "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(30, 30, 30)
	pdf.CellFormat(leftW, 5, tr(receipt.DeliveredTo), "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(rightW, 5, tr(receipt.ReceiptNumber), "", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(70, 70, 70)
	for _, l := range deliveryRecipientLines(receipt) {
		pdf.CellFormat(leftW, 4, tr(l), "", 0, "L", false, 0, "")
		pdf.CellFormat(rightW, 4, "", "", 1, "L", false, 0, "")
	}

	metaY := y
	pdf.SetXY(12+leftW, metaY+9)
	meta := [][2]string{
		{"Date:", receipt.ReceiptDate},
		{"Terms:", orDash(receipt.Terms)},
		{"P.O. No.:", orDash(receipt.PONumber)},
	}
	if v := strings.TrimSpace(receipt.InvoiceNumber); v != "" {
		meta = append(meta, [2]string{"Invoice:", v})
	}
	for _, mr := range meta {
		pdf.SetX(12 + leftW)
		pdf.CellFormat(22, 4, tr(mr[0]), "", 0, "L", false, 0, "")
		pdf.CellFormat(rightW-22, 4, tr(mr[1]), "", 1, "R", false, 0, "")
	}
	if pdf.GetY() < y+24 {
		pdf.SetY(y + 24)
	}
	pdf.Ln(2)
}

func deliveryRecipientLines(receipt DeliveryReceipt) []string {
	lines := make([]string, 0, 3)
	if v := strings.TrimSpace(receipt.TIN); v != "" {
		lines = append(lines, "TIN: "+v)
	}
	if v := strings.TrimSpace(receipt.Address); v != "" {
		lines = append(lines, v)
	}
	if v := strings.TrimSpace(receipt.RecipientEmail); v != "" {
		lines = append(lines, v)
	}
	return lines
}

func drawDeliveryReceiptLineItemsTable(pdf *fpdf.Fpdf, tr func(string) string, lines []DeliveryReceiptLine, pageWidth float64) {
	qtyW := pageWidth * 0.12
	unitW := pageWidth * 0.14
	descW := pageWidth - qtyW - unitW

	pdf.SetFillColor(240, 240, 240)
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetFont("Arial", "B", 7)
	pdf.SetTextColor(50, 50, 50)
	headers := []struct {
		label string
		w     float64
		align string
	}{
		{"QTY", qtyW, "C"},
		{"UNIT", unitW, "C"},
		{"ARTICLES", descW, "L"},
	}
	for _, h := range headers {
		pdf.CellFormat(h.w, 6, tr(h.label), "1", 0, h.align, true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(30, 30, 30)

	rowCount := deliveryReceiptTableRows
	if len(lines) > rowCount {
		rowCount = len(lines)
	}

	for i := 0; i < rowCount; i++ {
		qty, unit, desc := "", "", ""
		if i < len(lines) {
			qty = lines[i].Quantity
			unit = lines[i].Unit
			desc = lines[i].Description
		}

		x := pdf.GetX()
		y := pdf.GetY()
		pdf.MultiCell(descW, 5, tr(desc), "1", "L", false)
		rowH := pdf.GetY() - y
		if rowH < 5 {
			rowH = 5
		}
		pdf.SetXY(x, y)
		pdf.CellFormat(qtyW, rowH, tr(qty), "1", 0, "C", false, 0, "")
		pdf.CellFormat(unitW, rowH, tr(unit), "1", 0, "C", false, 0, "")
		pdf.SetXY(x+qtyW+unitW, y)
		pdf.CellFormat(descW, rowH, "", "1", 1, "L", false, 0, "")
	}
	pdf.Ln(2)
}

func drawDeliveryReceiptFooter(pdf *fpdf.Fpdf, tr func(string) string, config InvoiceConfig, pageWidth float64) {
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(60, 60, 60)
	ack := "Received the above articles in good order and condition."
	pdf.MultiCell(pageWidth, 4, tr(ack), "", "L", false)
	pdf.Ln(4)

	sigY := pdf.GetY()
	if sigY > 230 {
		pdf.AddPage()
		sigY = pdf.GetY()
	}
	colW := pageWidth / 3
	pdf.SetY(sigY)
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(80, 80, 80)
	for i, label := range []string{"Prepared by", "Delivered by", "Received by"} {
		x := 12 + float64(i)*colW
		pdf.SetXY(x, sigY+12)
		pdf.CellFormat(colW-4, 4, tr(label), "T", 0, "C", false, 0, "")
	}
	pdf.SetY(sigY + 20)
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

	disclaimer := "THIS DOCUMENT IS NOT VALID FOR CLAIM OF INPUT TAX."
	pdf.Ln(1)
	pdf.SetFont("Arial", "B", 7)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(pageWidth, 4, tr(disclaimer), "", 1, "C", false, 0, "")

	if strings.TrimSpace(config.FooterNotes) != "" {
		pdf.Ln(2)
		pdf.SetFont("Arial", "I", 7)
		pdf.SetTextColor(120, 120, 120)
		pdf.MultiCell(pageWidth, 3.5, tr(config.FooterNotes), "", "C", false)
	}
}

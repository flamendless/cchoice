package services

import (
	"bytes"
	"fmt"
	"strings"

	"cchoice/internal/constants"
	"cchoice/internal/enums"

	"github.com/go-pdf/fpdf"
)

const collectionReceiptSettlementRows = 11

func RenderCollectionReceiptPDF(config InvoiceConfig, receipt CollectionReceipt, settlements []CollectionReceiptSettlement) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 12)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	const pageWidth = 186.0
	const leftW = 62.0
	const rightW = pageWidth - leftW
	currency := cmpOr(config.Currency, constants.PHP)

	startY := pdf.GetY()
	logoRendered := drawInvoiceLogo(pdf, config, 12, startY)

	pdf.SetXY(12, startY)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(20, 20, 20)
	pdf.CellFormat(pageWidth, 7, tr("COLLECTION RECEIPT"), "", 1, "C", false, 0, "")

	if logoRendered {
		if y := startY + 22; pdf.GetY() < y {
			pdf.SetY(y)
		}
	}
	pdf.Ln(1)

	drawSellerHeader(pdf, tr, config, pageWidth)

	yMeta := pdf.GetY()
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(30, 30, 30)
	pdf.SetXY(12+leftW, yMeta)
	pdf.CellFormat(rightW*0.55, 5, tr("Receipt No."), "", 0, "L", false, 0, "")
	pdf.CellFormat(rightW*0.45, 5, tr(receipt.ReceiptNumber), "", 1, "R", false, 0, "")
	pdf.SetX(12 + leftW)
	pdf.CellFormat(rightW*0.55, 5, tr("Date:"), "", 0, "L", false, 0, "")
	pdf.CellFormat(rightW*0.45, 5, tr(receipt.ReceiptDate), "", 1, "R", false, 0, "")
	if pdf.GetY() < yMeta+12 {
		pdf.SetY(yMeta + 12)
	}
	pdf.Ln(2)

	bodyY := pdf.GetY()
	drawCollectionSettlementTable(pdf, tr, settlements, receipt.PaymentForm, currency, 12, bodyY, leftW)
	drawCollectionNarrativeBlock(pdf, tr, receipt, currency, 12+leftW, bodyY, rightW)

	bottomY := pdf.GetY()
	if leftBottom := bodyY + float64(collectionReceiptSettlementRows+6)*5 + 18; bottomY < leftBottom {
		pdf.SetY(leftBottom)
	}

	drawCollectionBIRFooter(pdf, tr, config, pageWidth)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawCollectionSettlementTable(pdf *fpdf.Fpdf, tr func(string) string, settlements []CollectionReceiptSettlement, paymentForm enums.ReceiptPaymentForm, currency string, x, y, w float64) {
	pdf.SetXY(x, y)
	pdf.SetFont("Arial", "B", 7)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(w, 4, tr("In Settlement of the following:"), "", 1, "L", false, 0, "")

	invW := w * 0.55
	amtW := w - invW
	pdf.SetFillColor(240, 240, 240)
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetFont("Arial", "B", 7)
	pdf.CellFormat(invW, 5, tr("Invoice No."), "1", 0, "C", true, 0, "")
	pdf.CellFormat(amtW, 5, tr("Amount"), "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(30, 30, 30)
	for i := 0; i < collectionReceiptSettlementRows; i++ {
		invNo := ""
		amt := ""
		if i < len(settlements) {
			invNo = settlements[i].InvoiceNumber
			amt = formatMoneyPlain(settlements[i].AmountRaw, currency)
		}
		pdf.SetX(x)
		pdf.CellFormat(invW, 5, tr(invNo), "1", 0, "L", false, 0, "")
		pdf.CellFormat(amtW, 5, amt, "1", 1, "R", false, 0, "")
	}

	pdf.Ln(2)
	drawCollectionPaymentForm(pdf, tr, paymentForm, x, pdf.GetY(), w)
}

func drawCollectionPaymentForm(pdf *fpdf.Fpdf, tr func(string) string, form enums.ReceiptPaymentForm, x, y, w float64) {
	pdf.SetXY(x, y)
	pdf.SetFont("Arial", "B", 7)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(w, 4, tr("Form of Payment"), "", 1, "L", false, 0, "")
	pdf.SetXY(x, y+5)
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(30, 30, 30)
	cashMark := "[ ]"
	checkMark := "[ ]"
	if form == enums.RECEIPT_PAYMENT_CASH {
		cashMark = "[X]"
	}
	if form == enums.RECEIPT_PAYMENT_CHECK {
		checkMark = "[X]"
	}
	pdf.CellFormat(w, 5, tr(fmt.Sprintf("%s Cash     %s Check", cashMark, checkMark)), "", 1, "L", false, 0, "")
}

func drawCollectionNarrativeBlock(pdf *fpdf.Fpdf, tr func(string) string, receipt CollectionReceipt, currency string, x, y, w float64) {
	pdf.SetXY(x, y)
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(30, 30, 30)

	receivedFrom := orDash(receipt.ReceivedFrom)
	tin := orDash(receipt.TIN)
	address := orDash(receipt.Address)
	amountWords := orDash(receipt.AmountInWords)
	amountFig := formatMoneyPlain(receipt.AmountRaw, currency)
	paymentFor := orDash(receipt.PaymentFor)

	pdf.MultiCell(w, 4.5, tr(fmt.Sprintf("Received from %s with TIN %s", receivedFrom, tin)), "", "L", false)
	pdf.SetX(x)
	pdf.MultiCell(w, 4.5, tr("and address at "+address), "", "L", false)
	pdf.SetX(x)
	pdf.MultiCell(w, 4.5, tr(amountWords+", the sum of"), "", "L", false)
	pdf.SetX(x)
	pdf.MultiCell(w, 4.5, tr("(P "+strings.TrimPrefix(amountFig, currency+" ")+") in partial/full payment for "+paymentFor), "", "L", false)
	pdf.Ln(4)

	drawCollectionOSCAPWDGrid(pdf, tr, receipt, x, pdf.GetY(), w)
	pdf.Ln(2)

	pdf.SetX(x)
	pdf.SetFont("Arial", "B", 7)
	pdf.SetTextColor(40, 40, 40)
	pdf.MultiCell(w, 4, tr("THIS DOCUMENT IS NOT VALID FOR CLAIM OF INPUT TAXES."), "", "C", false)
	pdf.Ln(4)

	sigY := pdf.GetY()
	pdf.SetXY(x, sigY)
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(w, 4, tr("By:"), "", 1, "L", false, 0, "")
	pdf.SetX(x)
	pdf.CellFormat(w, 8, "", "B", 1, "L", false, 0, "")
	pdf.SetX(x)
	pdf.CellFormat(w, 4, tr("Cashier/Authorized Representative"), "", 1, "C", false, 0, "")
}

func drawCollectionOSCAPWDGrid(pdf *fpdf.Fpdf, tr func(string) string, receipt CollectionReceipt, x, y, w float64) {
	colW := w / 2
	rowH := 8.0
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetFont("Arial", "B", 6)
	pdf.SetTextColor(80, 80, 80)

	pdf.SetXY(x, y)
	pdf.CellFormat(w, 4, tr("Sr. Citizen TIN"), "1", 1, "L", false, 0, "")
	pdf.SetX(x)
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(30, 30, 30)
	pdf.CellFormat(w, rowH, tr(orDash(receipt.SCCitizenTIN)), "1", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 6)
	pdf.SetTextColor(80, 80, 80)
	pdf.SetX(x)
	pdf.CellFormat(colW, 4, tr("OSCA/PWD ID No."), "1", 0, "L", false, 0, "")
	pdf.CellFormat(colW, 4, tr("Signature"), "1", 1, "L", false, 0, "")
	pdf.SetX(x)
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(30, 30, 30)
	pdf.CellFormat(colW, rowH, tr(orDash(receipt.OSCAPWDIDNo)), "1", 0, "L", false, 0, "")
	pdf.CellFormat(colW, rowH, "", "1", 1, "L", false, 0, "")
}

func drawCollectionBIRFooter(pdf *fpdf.Fpdf, tr func(string) string, config InvoiceConfig, pageWidth float64) {
	pdf.Ln(4)
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
}

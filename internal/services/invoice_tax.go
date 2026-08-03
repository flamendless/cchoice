package services

import (
	"math"

	"cchoice/internal/enums"
)

type BIRTaxLineInput struct {
	LineTotal int64
	TaxType   enums.InvoiceLineTaxType
}

type BIRTaxAdjustments struct {
	WithholdingTax int64
	SCPWDDiscount  int64
	AddVAT         int64
}

type BIRTaxSummary struct {
	VATableSales           int64
	VATExemptSales         int64
	ZeroRatedSales         int64
	VATAmount              int64
	TotalSales             int64
	TotalSalesVATInclusive int64
	LessVAT                int64
	WithholdingTax         int64
	AmountNetOfVAT         int64
	SCPWDDiscount          int64
	AddVAT                 int64
	Total                  int64
	Subtotal               int64
}

func ComputeBIRTaxSummary(lines []BIRTaxLineInput, vatRatePct float64, adj BIRTaxAdjustments) BIRTaxSummary {
	var vatableNet, vatExempt, zeroRated, vatAmount int64
	for _, ln := range lines {
		switch ln.TaxType {
		case enums.INVOICE_LINE_TAX_VAT_EXEMPT:
			vatExempt += ln.LineTotal
		case enums.INVOICE_LINE_TAX_ZERO_RATED:
			zeroRated += ln.LineTotal
		default:
			if vatRatePct > 0 {
				net := int64(math.Round(float64(ln.LineTotal) / (1.0 + vatRatePct/100.0)))
				vatableNet += net
				vatAmount += ln.LineTotal - net
			} else {
				vatableNet += ln.LineTotal
			}
		}
	}

	totalVATInclusive := vatableNet + vatExempt + zeroRated + vatAmount
	totalSales := vatableNet + vatExempt + zeroRated
	lessVAT := vatAmount
	amountNet := totalVATInclusive - lessVAT - adj.WithholdingTax
	total := amountNet - adj.SCPWDDiscount + adj.AddVAT

	return BIRTaxSummary{
		VATableSales:           vatableNet,
		VATExemptSales:         vatExempt,
		ZeroRatedSales:         zeroRated,
		VATAmount:              vatAmount,
		TotalSales:             totalSales,
		TotalSalesVATInclusive: totalVATInclusive,
		LessVAT:                lessVAT,
		WithholdingTax:         adj.WithholdingTax,
		AmountNetOfVAT:         amountNet,
		SCPWDDiscount:          adj.SCPWDDiscount,
		AddVAT:                 adj.AddVAT,
		Total:                  total,
		Subtotal:               totalVATInclusive,
	}
}

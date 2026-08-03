package services

import (
	"testing"

	"cchoice/internal/enums"

	"github.com/stretchr/testify/assert"
)

func TestComputeBIRTaxSummary_VATInclusive(t *testing.T) {
	lines := []BIRTaxLineInput{
		{LineTotal: 112000, TaxType: enums.INVOICE_LINE_TAX_VATABLE},
		{LineTotal: 50000, TaxType: enums.INVOICE_LINE_TAX_VAT_EXEMPT},
	}
	summary := ComputeBIRTaxSummary(lines, 12, BIRTaxAdjustments{})

	assert.Equal(t, int64(100000), summary.VATableSales)
	assert.Equal(t, int64(12000), summary.VATAmount)
	assert.Equal(t, int64(50000), summary.VATExemptSales)
	assert.Equal(t, int64(150000), summary.TotalSales)
	assert.Equal(t, int64(162000), summary.TotalSalesVATInclusive)
	assert.Equal(t, int64(150000), summary.Total)
}

func TestComputeBIRTaxSummary_WithAdjustments(t *testing.T) {
	lines := []BIRTaxLineInput{
		{LineTotal: 112000, TaxType: enums.INVOICE_LINE_TAX_VATABLE},
	}
	summary := ComputeBIRTaxSummary(lines, 12, BIRTaxAdjustments{
		WithholdingTax: 2000,
		SCPWDDiscount:  1000,
		AddVAT:         500,
	})

	assert.Equal(t, int64(98000), summary.AmountNetOfVAT)
	assert.Equal(t, int64(97500), summary.Total)
}

func TestComputeBIRTaxSummary_ZeroRated(t *testing.T) {
	lines := []BIRTaxLineInput{
		{LineTotal: 25000, TaxType: enums.INVOICE_LINE_TAX_ZERO_RATED},
	}
	summary := ComputeBIRTaxSummary(lines, 12, BIRTaxAdjustments{})
	assert.Equal(t, int64(25000), summary.ZeroRatedSales)
	assert.Equal(t, int64(25000), summary.Total)
}

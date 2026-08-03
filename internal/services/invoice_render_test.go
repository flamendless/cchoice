package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildEmailTemplateData_CustomerTotal(t *testing.T) {
	svc := &InvoiceService{}
	invoice := Invoice{
		InvoiceNumber:             "INV-001",
		Currency:                  "PHP",
		VATPercentage:             "12",
		TotalSalesRaw:             150000,
		VATAmountRaw:              18000,
		TotalSalesVATInclusiveRaw: 168000,
		TotalRaw:                  140000, // after withholding / adjustments
		SubtotalRaw:               168000,
	}
	data := svc.BuildEmailTemplateData(InvoiceConfig{}, invoice, nil)

	assert.Equal(t, "₱1,500.00", data["Subtotal"])
	assert.Equal(t, "₱1,680.00", data["Total"])
}

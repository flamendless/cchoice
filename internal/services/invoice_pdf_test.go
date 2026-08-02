package services

import (
	"testing"

	"cchoice/internal/enums"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatMoneyPlain(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		want     string
		centavos int64
	}{
		{"zero", "PHP", "PHP 0.00", 0},
		{"simple", "PHP", "PHP 1,500.00", 150000},
		{"with centavos", "PHP", "PHP 99.50", 9950},
		{"thousands", "PHP", "PHP 1,234,567.89", 123456789},
		{"negative", "PHP", "PHP -12.34", -1234},
		{"no currency", "", "10.00", 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatMoneyPlain(tt.centavos, tt.currency))
		})
	}
}

func TestRenderInvoicePDF(t *testing.T) {
	config := InvoiceConfig{
		BusinessName:  "Test Business Inc.",
		Address:       "123 Test Street, Manila",
		TIN:           "000-111-222-333",
		Email:         "billing@test.com",
		ContactNumber: "+63 900 000 0000",
		Currency:      "PHP",
		VATPercentage: "12",
		FooterNotes:   "Thank you for your business!",
	}
	invoice := Invoice{
		InvoiceNumber:  "INV-20260802-00001",
		Status:         enums.INVOICE_STATUS_ISSUED,
		RecipientName:  "Acme Corp",
		RecipientEmail: "ap@acme.com",
		IssueDate:      "2026-08-02",
		DueDate:        "2026-08-16",
		Currency:       "PHP",
		VATPercentage:  "12",
		SubtotalRaw:    150000,
		VATAmountRaw:   18000,
		TotalRaw:       168000,
		Notes:          "Delivery within 7 days.",
	}
	lines := []InvoiceLine{
		{Description: "Cement bag", Quantity: 10, UnitPriceRaw: 10000, LineTotalRaw: 100000},
		{Description: "Steel bar", Quantity: 5, UnitPriceRaw: 10000, LineTotalRaw: 50000},
	}

	data, err := RenderInvoicePDF(config, invoice, lines)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	// PDF files start with the "%PDF" magic bytes.
	assert.Equal(t, "%PDF", string(data[:4]))
}

package services

import (
	"testing"

	"cchoice/internal/enums"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderDeliveryReceiptPDF(t *testing.T) {
	t.Parallel()
	config := InvoiceConfig{
		BusinessName: "C-Choice Construction Supply",
		Address:      "General Trias, Cavite",
		TIN:          "319-079-964-00000",
		Currency:     "PHP",
	}
	receipt := DeliveryReceipt{
		ReceiptNumber:  "DR-20260115-00001",
		DeliveredTo:    "Test Customer",
		RecipientEmail: "test@example.com",
		TIN:            "123-456-789-000",
		Address:        "Test Address",
		ReceiptDate:    "2026-01-15",
		Terms:          "Net 30",
		PONumber:       "PO-001",
		Status:         enums.RECEIPT_STATUS_ISSUED,
	}
	lines := []DeliveryReceiptLine{
		{Quantity: "10", Unit: "pcs", Description: "Cement"},
	}
	pdf, err := RenderDeliveryReceiptPDF(config, receipt, lines)
	require.NoError(t, err)
	assert.NotEmpty(t, pdf)
}

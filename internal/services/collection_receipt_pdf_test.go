package services

import (
	"testing"

	"cchoice/internal/enums"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderCollectionReceiptPDF(t *testing.T) {
	t.Parallel()
	config := InvoiceConfig{
		BusinessName: "C-Choice Construction Supply",
		Address:      "General Trias, Cavite",
		TIN:          "319-079-964-00000",
		Currency:     "PHP",
	}
	receipt := CollectionReceipt{
		ReceiptNumber:  "CR-20260115-00001",
		ReceivedFrom:   "Test Customer",
		RecipientEmail: "test@example.com",
		TIN:            "123-456-789-000",
		Address:        "Test Address",
		ReceiptDate:    "2026-01-15",
		Amount:         "PHP 1,000.00",
		AmountRaw:      100000,
		AmountInWords:  "One Thousand Pesos",
		PaymentFor:     "Materials",
		PaymentForm:    enums.RECEIPT_PAYMENT_CASH,
		Status:         enums.RECEIPT_STATUS_ISSUED,
	}
	settlements := []CollectionReceiptSettlement{
		{InvoiceNumber: "INV-001", Amount: "PHP 1,000.00", AmountRaw: 100000},
	}
	pdf, err := RenderCollectionReceiptPDF(config, receipt, settlements)
	require.NoError(t, err)
	assert.NotEmpty(t, pdf)
}

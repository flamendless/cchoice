package errs

import "errors"

var (
	ErrDeliveryReceipt                   = errors.New("[DELIVERY_RECEIPT]: Error on delivery receipt service")
	ErrDeliveryReceiptNotFound           = errors.New("[DELIVERY_RECEIPT]: Delivery receipt not found")
	ErrDeliveryReceiptNoLines            = errors.New("[DELIVERY_RECEIPT]: At least one line item is required")
	ErrDeliveryReceiptCreateFailed       = errors.New("[DELIVERY_RECEIPT]: Failed to create delivery receipt")
	ErrDeliveryReceiptPDFFailed          = errors.New("[DELIVERY_RECEIPT]: Failed to generate delivery receipt PDF")
	ErrDeliveryReceiptEmailFailed        = errors.New("[DELIVERY_RECEIPT]: Failed to send delivery receipt email")
	ErrDeliveryReceiptEmailNotConfigured = errors.New("[DELIVERY_RECEIPT]: Email service is not configured")
	ErrDeliveryReceiptRecipientNoEmail   = errors.New("[DELIVERY_RECEIPT]: Recipient has no email address")
)

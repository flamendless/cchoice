package errs

import "errors"

var (
	ErrCollectionReceipt                   = errors.New("[COLLECTION_RECEIPT]: Error on collection receipt service")
	ErrCollectionReceiptNotFound           = errors.New("[COLLECTION_RECEIPT]: Collection receipt not found")
	ErrCollectionReceiptCreateFailed       = errors.New("[COLLECTION_RECEIPT]: Failed to create collection receipt")
	ErrCollectionReceiptPDFFailed          = errors.New("[COLLECTION_RECEIPT]: Failed to generate collection receipt PDF")
	ErrCollectionReceiptEmailFailed        = errors.New("[COLLECTION_RECEIPT]: Failed to send collection receipt email")
	ErrCollectionReceiptEmailNotConfigured = errors.New("[COLLECTION_RECEIPT]: Email service is not configured")
	ErrCollectionReceiptRecipientNoEmail   = errors.New("[COLLECTION_RECEIPT]: Recipient has no email address")
	ErrCollectionReceiptSettlementMismatch = errors.New("[COLLECTION_RECEIPT]: Receipt amount must match the sum of settlement rows")
)

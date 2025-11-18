package errs

import "errors"

var (
	ErrReceiptInvalidImage   = errors.New("[RECEIPT]: Invalid image file")
	ErrReceiptImageNotFound  = errors.New("[RECEIPT]: Image file not found")
	ErrReceiptNoTextFound    = errors.New("[RECEIPT]: No text found in image")
	ErrReceiptParsingFailed  = errors.New("[RECEIPT]: Failed to parse receipt data")
	ErrReceiptWriteFailed    = errors.New("[RECEIPT]: Failed to write output")
	ErrReceiptInvalidFormat  = errors.New("[RECEIPT]: Invalid output format")
)

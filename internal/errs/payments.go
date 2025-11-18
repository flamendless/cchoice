package errs

import "errors"

var (
	ErrPaymentPayload  = errors.New("[PAYMENT]: Processing payload failed")
	ErrPaymentClient   = errors.New("[PAYMENT]: HTTP client")
	ErrPaymentResponse = errors.New("[PAYMENT]: Processing response failed")
)

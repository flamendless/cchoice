package errs

import "errors"

var (
	ErrPaymentPayload  = errors.New("[PAYMENT]: Error in processing payload")
	ErrPaymentClient   = errors.New("[PAYMENT]: Error in HTTP client")
	ErrPaymentResponse = errors.New("[PAYMENT]: Error in processing response")
)

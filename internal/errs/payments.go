package errs

import "errors"

var (
	ErrPaymentPayload  = errors.New("[PAYMENT]: Error in processing payload")
	ErrPaymenetClient  = errors.New("[PAYMENT]: Error in HTTP client")
	ErrPaymentResponse = errors.New("[PAYMENT]: Error in processing response")
)

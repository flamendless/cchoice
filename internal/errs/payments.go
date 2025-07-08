package errs

import "errors"

var (
	ERR_PAYMENT_PAYLOAD  = errors.New("[PAYMENT]: Error in processing payload")
	ERR_PAYMENT_CLIENT   = errors.New("[PAYMENT]: Error in HTTP client")
	ERR_PAYMENT_RESPONSE = errors.New("[PAYMENT]: Error in processing response")
)

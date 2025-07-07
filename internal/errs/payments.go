package errs

import "errors"

var (
	ERR_PAYMENT_PAYLOAD  = errors.New("Error in processing payload")
	ERR_PAYMENT_CLIENT   = errors.New("Error in HTTP client")
	ERR_PAYMENT_RESPONSE = errors.New("Error in processing response")
)

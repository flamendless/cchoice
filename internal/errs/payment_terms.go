package errs

import "errors"

var (
	ErrInvalidPaymentTerms     = errors.New("[PAYMENT_TERMS]: Invalid payment terms")
	ErrInvalidPaymentTermsUnit = errors.New("[PAYMENT_TERMS]: Invalid payment terms unit")
)

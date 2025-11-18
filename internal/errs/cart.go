package errs

import "errors"

var (
	ErrCartMissingCheckoutLines = errors.New("[CART]: No checkout lines found for the given session")
)

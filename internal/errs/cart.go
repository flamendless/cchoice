package errs

import "errors"

var (
	ErrCartMissingCheckoutLines = errors.New("[Cart]: No checkout lines found for the given session")
)

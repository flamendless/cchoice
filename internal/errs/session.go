package errs

import "errors"

var (
	ErrSessionCheckoutLineProductIDs = errors.New("[Session]: Failed to cast product IDs to []string")
)

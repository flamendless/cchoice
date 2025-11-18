package errs

import "errors"

var (
	ErrSessionCheckoutLineProductIDs = errors.New("[SESSION]: Failed to cast product IDs to []string")
)

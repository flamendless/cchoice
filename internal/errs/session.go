package errs

import "errors"

var (
	ERR_SESSION_CHECKOUT_LINE_PRODUCT_IDS = errors.New("[Session]: Failed to cast product IDs to []string")
)

package errs

import "errors"

var (
	ErrPaymongoServiceInit    = errors.New("[PAYMONGO]: Service must be configured")
	ErrPaymongoAPIKeyRequired = errors.New("[PAYMONGO]: API key required")
)

package errs

import "errors"

var (
	ErrGVisionServiceInit    = errors.New("[GVISION]: Service must be configured")
	ErrGVisionAPIKeyRequired = errors.New("[GVISION]: API key required")
	ErrGVisionAPI            = errors.New("[GVISION]: Google Vision API error")
)

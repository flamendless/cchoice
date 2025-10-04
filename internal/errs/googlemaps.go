package errs

import "errors"

var (
	ErrGMapsAPIKeyRequired  = errors.New("[GMAPS]: API key is required")
	ErrGMapsInvalidResponse = errors.New("[GMAPS]: Invalid response")
	ErrGMapsNoResults       = errors.New("[GMAPS]: No results found")
	ErrGMapsQuotaExceeded   = errors.New("[GMAPS]: API quota exceeded")
	ErrGMapsInvalidRequest  = errors.New("[GMAPS]: Invalid request")
	ErrGMapsRequestDenied   = errors.New("[GMAPS]: Request denied")
	ErrGMapsUnknownError    = errors.New("[GMAPS]: Unknown error")
)

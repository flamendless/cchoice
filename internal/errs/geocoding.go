package errs

import "errors"

var (
	ErrAPIKeyRequired  = errors.New("google Maps API key is required")
	ErrInvalidResponse = errors.New("invalid response from Google Maps API")
	ErrNoResults       = errors.New("no results found")
	ErrQuotaExceeded   = errors.New("API quota exceeded")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrRequestDenied   = errors.New("request denied")
	ErrUnknownError    = errors.New("unknown error from Google Maps API")
)

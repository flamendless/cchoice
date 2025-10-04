package errs

import "errors"

var (
	ErrSign        = errors.New("[Signing]: Error in signing request")
	ErrHTTPRequest = errors.New("[HTTP Request]")
)

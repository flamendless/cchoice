package errs

import "errors"

var (
	ErrSign        = errors.New("[SIGNING]")
	ErrHTTPRequest = errors.New("[HTTP REQ]")
)

package errs

import "errors"

var (
	ErrInvalidParams     = errors.New("[CLIENT]: Invalid params")
	ErrInvalidParamLimit = errors.New("[CLIENT]: Invalid 'limit' param")
)

package errs

import "errors"

var (
	ERR_INVALID_PARAMS      = errors.New("[CLIENT]: Invalid params")
	ERR_INVALID_PARAM_LIMIT = errors.New("[CLIENT]: Invalid 'limit' param")
)

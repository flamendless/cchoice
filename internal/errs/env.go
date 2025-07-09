package errs

import "errors"

var (
	ERR_ENV_VAR_REQUIRED = errors.New("[Env]: Required")
)

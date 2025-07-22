package errs

import "errors"

var (
	ErrEnvVarRequired = errors.New("[Env]: Required")
)

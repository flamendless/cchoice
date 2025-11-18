package errs

import "errors"

var (
	ErrEnvVarRequired = errors.New("[ENV]: Required")
)

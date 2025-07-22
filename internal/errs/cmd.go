package errs

import "errors"

var (
	ErrCmd         = errors.New("[CMD]: Error in command")
	ErrCmdRequired = errors.New("[CMD]: Required flag")
)

package errs

import "errors"

var (
	ERR_CMD          = errors.New("[CMD]: Error in command")
	ERR_CMD_REQUIRED = errors.New("[CMD]: Required flag")
)

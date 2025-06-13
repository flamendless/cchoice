package errs

import "errors"

var (
	ERR_CMD          = errors.New("Error in command")
	ERR_CMD_REQUIRED = errors.New("Required flag")
)

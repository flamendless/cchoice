package errs

import "errors"

var (
	ErrServerInit = errors.New("[Server]: Failed to initialize server")
)

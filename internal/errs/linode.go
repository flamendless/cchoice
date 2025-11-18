package errs

import "errors"

var (
	ErrLinodeServiceInit   = errors.New("[LINODE]: Service must be configured")
)

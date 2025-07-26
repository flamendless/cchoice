package errs

import "errors"

var (
	ErrFS          = errors.New("[FS]: Error on filesystem")
	ErrImagePrefix = errors.New("[FS]: Invalid image prefix")
)

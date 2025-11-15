package errs

import "errors"

var (
	ErrFS               = errors.New("[FS]: Error on filesystem")
	ErrImagePrefix      = errors.New("[FS]: Invalid image prefix")
	ErrNotADirectory    = errors.New("[FS]: Not a directory")
	ErrSeekNotSupported = errors.New("[FS]: Seek not supported")
)

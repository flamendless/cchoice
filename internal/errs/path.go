package errs

import "errors"

var (
	ErrPathEmpty            = errors.New("[PATH]: Empty path")
	ErrPathTraversalAttempt = errors.New("[PATH]: Traversal attempt detected")
	ErrPathPrefix           = errors.New("[PATH]: Invalid prefix")
	ErrPathTooLong          = errors.New("[PATH]: Too long")
	ErrPathInvalidExt       = errors.New("[PATH]: Invalid extension")
)

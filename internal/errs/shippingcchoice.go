package errs

import "errors"

var (
	ErrCChoice            = errors.New("[CCHOICE]")
	ErrCChoiceServiceInit = errors.New("[CCHOICE]: Service must be configured")
)

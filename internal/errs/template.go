package errs

import "errors"

var (
	ErrTemplateRead    = errors.New("[TEMPLATE]: Failed to read")
	ErrTemplateParse   = errors.New("[TEMPLATE]: Failed to parse")
	ErrTemplateExecute = errors.New("[TEMPLATE]: Failed to execute")
)

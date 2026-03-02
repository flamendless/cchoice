package errs

import "errors"

var (
	ErrSettingsRequired = errors.New("[Settings]: Required")
)

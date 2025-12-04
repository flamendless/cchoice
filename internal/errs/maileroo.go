package errs

import "errors"

var (
	ErrMailerooServiceInit    = errors.New("[MAILEROO]: Service must be configured")
	ErrMailerooAPIKeyRequired = errors.New("[MAILEROO]: API key required")
	ErrMailerooFromRequired   = errors.New("[MAILEROO]: From required")
	ErrMailerooSendFailed     = errors.New("[MAILEROO]: Failed to send email")
)

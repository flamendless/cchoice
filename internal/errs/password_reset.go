package errs

import "errors"

var (
	ErrPasswordResetRateLimited = errors.New("[PASSWORD_RESET]: Rate limited, please wait before requesting another reset")
	ErrInvalidResetToken        = errors.New("[PASSWORD_RESET]: Invalid or expired reset token")
	ErrPasswordResetFailed      = errors.New("[PASSWORD_RESET]: Failed to reset password")
	ErrResetTokenGeneration     = errors.New("[PASSWORD_RESET]: Failed to generate reset token")
)

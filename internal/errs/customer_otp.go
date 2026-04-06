package errs

import "errors"

var (
	ErrOTPGenerationFailed = errors.New("[OTP]: Failed to generate OTP code")
	ErrOTPCreationFailed   = errors.New("[OTP]: Failed to create OTP code")
	ErrInvalidOTP          = errors.New("[OTP]: Invalid or expired OTP code")
	ErrOTPRateLimited      = errors.New("[OTP]: Rate limited, please wait before requesting another code")
	ErrOTPUpdateFailed     = errors.New("[OTP]: Failed to update OTP status")

	ErrCustomerStatusUpdateFailed = errors.New("[CUSTOMER]: Failed to update customer status")
)

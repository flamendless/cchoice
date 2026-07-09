package errs

import "errors"

var (
	ErrCustomerProfileUpdateFailed  = errors.New("[CUSTOMER]: Failed to update profile")
	ErrCustomerPasswordVerifyFailed = errors.New("[CUSTOMER]: Unable to verify current password")
	ErrCustomerPasswordIncorrect    = errors.New("[CUSTOMER]: Current password is incorrect")
	ErrCustomerPasswordUpdateFailed = errors.New("[CUSTOMER]: Failed to update password")
	ErrCustomerOTPUnableToSend      = errors.New("[CUSTOMER]: Unable to send verification code")
)

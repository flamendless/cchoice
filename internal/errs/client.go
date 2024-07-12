package errs

import "errors"

var (
	ERR_PARSE_FORM              = errors.New("Failed to parse form")
	ERR_NO_AUTH                 = errors.New("Not authenticated")
	ERR_ALREADY_OTP_ENROLLED    = errors.New("User is already enrolled in OTP")
	ERR_CHOOSE_VALID_OPTION     = errors.New("Please choose a valid option")
	ERR_EXPIRED_OTP_LOGIN_AGAIN = errors.New("Expired OTP session. Log in again")
)

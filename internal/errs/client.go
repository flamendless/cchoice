package errs

import "errors"

var (
	ERR_INVALID_INPUT           = errors.New("Invalid input")
	ERR_INVALID_RESOURCE        = errors.New("Invalid resource")
	ERR_INVALID_PARAMS          = errors.New("Invalid params")
	ERR_INVALID_CREDENTIALS     = errors.New("Invalid credentials")
	ERR_PARSE_FORM              = errors.New("Failed to parse form")
	ERR_NO_AUTH                 = errors.New("Not authenticated")
	ERR_NEED_OTP                = errors.New("Need OTP")
	ERR_INVALID_OTP             = errors.New("Invalid OTP")
	ERR_INVALID_TOKEN           = errors.New("Invalid token")
	ERR_ALREADY_OTP_ENROLLED    = errors.New("User is already enrolled in OTP")
	ERR_CHOOSE_VALID_OPTION     = errors.New("Please choose a valid option")
	ERR_EXPIRED_OTP_LOGIN_AGAIN = errors.New("Expired OTP session. Log in again")
	ERR_EXPIRED_REGISTRATION    = errors.New("Expired session. Register again")
	ERR_EXPIRED_SESSION         = errors.New("Expired session. Log in again")
)

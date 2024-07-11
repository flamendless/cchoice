package errs

import "errors"

var (
	ERR_NO_AUTH              = errors.New("Not authenticated")
	ERR_ALREADY_OTP_ENROLLED = errors.New("User is already enrolled in OTP")
	ERR_CHOOSE_VALID_OPTION  = errors.New("Please choose a valid option")
)

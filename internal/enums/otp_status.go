package enums

//go:generate stringer -type=OTPStatus -trimprefix=OTP_STATUS_

import (
	"fmt"
)

type OTPStatus int

const (
	OTP_STATUS_UNDEFINED OTPStatus = iota
	OTP_STATUS_INITIAL
	OTP_STATUS_ENROLLED
	OTP_STATUS_SENT_CODE
	OTP_STATUS_VALID
)

func ParseOTPStatusEnum(e string) OTPStatus {
	switch e {
	case OTP_STATUS_UNDEFINED.String():
		return OTP_STATUS_UNDEFINED
	case OTP_STATUS_INITIAL.String():
		return OTP_STATUS_INITIAL
	case OTP_STATUS_ENROLLED.String():
		return OTP_STATUS_ENROLLED
	case OTP_STATUS_SENT_CODE.String():
		return OTP_STATUS_SENT_CODE
	case OTP_STATUS_VALID.String():
		return OTP_STATUS_VALID
	default:
		panic(fmt.Sprintf("Can't convert '%s' to OTPStatus enum", e))
	}
}

package enums

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

func (t OTPStatus) String() string {
	switch t {
	case OTP_STATUS_UNDEFINED:
		return "UNDEFINED"
	case OTP_STATUS_INITIAL:
		return "INITIAL"
	case OTP_STATUS_ENROLLED:
		return "ENROLLED"
	case OTP_STATUS_SENT_CODE:
		return "SENT_CODE"
	case OTP_STATUS_VALID:
		return "VALID"
	default:
		panic("unknown enum")
	}
}

func ParseOTPStatusEnum(e string) OTPStatus {
	switch e {
	case "UNDEFINED":
		return OTP_STATUS_UNDEFINED
	case "INITIAL":
		return OTP_STATUS_INITIAL
	case "ENROLLED":
		return OTP_STATUS_ENROLLED
	case "SENT_CODE":
		return OTP_STATUS_SENT_CODE
	case "VALID":
		return OTP_STATUS_VALID
	default:
		panic(fmt.Sprintf("Can't convert '%s' to OTPStatus enum", e))
	}
}

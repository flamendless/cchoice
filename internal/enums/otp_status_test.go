package enums

import (
	"testing"
)

var tblOTPStatus = map[OTPStatus]string{
	OTP_STATUS_UNDEFINED: "UNDEFINED",
	OTP_STATUS_INITIAL:   "INITIAL",
	OTP_STATUS_ENROLLED:  "ENROLLED",
	OTP_STATUS_SENT_CODE: "SENT_CODE",
	OTP_STATUS_VALID:     "VALID",
}

func TestOTPStatusToString(t *testing.T) {
	for otpstatus, val := range tblOTPStatus {
		if otpstatus.String() != val {
			t.Fatalf("Mismatch: %s = %s", otpstatus.String(), val)
		}
	}
}

func TestParseOTPStatusEnum(t *testing.T) {
	for otpstatus, val := range tblOTPStatus {
		parsed := ParseOTPStatusEnum(val)
		if parsed != otpstatus {
			t.Fatalf("Mismatch: %s = %s", val, otpstatus)
		}
	}
}

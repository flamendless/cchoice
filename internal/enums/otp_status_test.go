package enums

import (
	"testing"

	"github.com/stretchr/testify/require"
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
		t.Run(val, func(t *testing.T) {
			require.Equal(t, val, otpstatus.String())
		})
	}
}

func TestParseOTPStatusEnum(t *testing.T) {
	for otpstatus, val := range tblOTPStatus {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, otpstatus, ParseOTPStatusEnum(val))
		})
	}
}

func BenchmarkOTPStatusToString(b *testing.B) {
	for otpstatus := range tblOTPStatus {
		for i := 0; i < b.N; i++ {
			_ = otpstatus.String()
		}
	}
}

func BenchmarkParseOTPStatusEnum(b *testing.B) {
	for _, val := range tblOTPStatus {
		for i := 0; i < b.N; i++ {
			_ = ParseOTPStatusEnum(val)
		}
	}
}

package enums

import (
	pb "cchoice/proto"
	"testing"

	"github.com/stretchr/testify/require"
)

var tblOTPMethod = map[string]pb.OTPMethod_OTPMethod{
	"UNDEFINED":     pb.OTPMethod_UNDEFINED,
	"AUTHENTICATOR": pb.OTPMethod_AUTHENTICATOR,
	"SMS":           pb.OTPMethod_SMS,
	"EMAIL":         pb.OTPMethod_EMAIL,
}

func TestParseOTPMethodEnumPB(t *testing.T) {
	for val, otp := range tblOTPMethod {
		t.Run(val, func(t *testing.T) {
			enum := StringToPBEnum(
				val,
				pb.OTPMethod_OTPMethod_value,
				pb.OTPMethod_UNDEFINED,
			)
			require.Equal(t, otp, enum)
		})
	}
}

func BenchmarkParseOTPMethodEnumPB(b *testing.B) {
	for val := range tblOTPMethod {
		for i := 0; i < b.N; i++ {
			_ = StringToPBEnum(
				val,
				pb.OTPMethod_OTPMethod_value,
				pb.OTPMethod_UNDEFINED,
			)
		}
	}
}

package enums

import (
	pb "cchoice/proto"
	"testing"
)

var tblOTPMethod = map[string]pb.OTPMethod_OTPMethod{
	"UNDEFINED": pb.OTPMethod_UNDEFINED,
	"AUTHENTICATOR": pb.OTPMethod_AUTHENTICATOR,
	"SMS": pb.OTPMethod_SMS,
	"EMAIL": pb.OTPMethod_EMAIL,
}

func TestParseOTPMethodEnumPB(t *testing.T) {
	for val, otp := range tblOTPMethod {
		enum := StringToPBEnum(
			val,
			pb.OTPMethod_OTPMethod_value,
			pb.OTPMethod_UNDEFINED,
		)
		if enum != otp {
			t.Fatalf("Mismatch: %s = %s", enum, val)
		}
	}
}

package enums

import (
	pb "cchoice/proto"
	"testing"
)

func TestParseOTPMethodEnumPB(t *testing.T) {
	undef := StringToPBEnum(
		"UNDEFINED",
		pb.OTPMethod_OTPMethod_value,
		pb.OTPMethod_UNDEFINED,
	)
	authenticator := StringToPBEnum(
		"AUTHENTICATOR",
		pb.OTPMethod_OTPMethod_value,
		pb.OTPMethod_UNDEFINED,
	)
	sms := StringToPBEnum(
		"SMS",
		pb.OTPMethod_OTPMethod_value,
		pb.OTPMethod_UNDEFINED,
	)
	email := StringToPBEnum(
		"EMAIL",
		pb.OTPMethod_OTPMethod_value,
		pb.OTPMethod_UNDEFINED,
	)

	if undef != pb.OTPMethod_UNDEFINED {
		t.Fatalf("Mismatch: %s = %s", undef, pb.OTPMethod_UNDEFINED)
	}
	if authenticator != pb.OTPMethod_AUTHENTICATOR {
		t.Fatalf("Mismatch: %s = %s", authenticator, pb.OTPMethod_AUTHENTICATOR)
	}
	if sms != pb.OTPMethod_SMS {
		t.Fatalf("Mismatch: %s = %s", sms, pb.OTPMethod_SMS)
	}
	if email != pb.OTPMethod_EMAIL {
		t.Fatalf("Mismatch: %s = %s", email, pb.OTPMethod_EMAIL)
	}
}

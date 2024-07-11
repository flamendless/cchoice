package enums

import (
	"testing"
)

func TestOTPStatusToString(t *testing.T) {
	undef := OTP_STATUS_UNDEFINED
	initial := OTP_STATUS_INITIAL
	enrolled := OTP_STATUS_ENROLLED
	sentCode := OTP_STATUS_SENT_CODE
	valid := OTP_STATUS_VALID

	if undef.String() != "UNDEFINED" {
		t.Fatalf("Mismatch: %s = %s", undef.String(), "UNDEFINED")
	}
	if initial.String() != "INITIAL" {
		t.Fatalf("Mismatch: %s = %s", initial.String(), "INITIAL")
	}
	if enrolled.String() != "ENROLLED" {
		t.Fatalf("Mismatch: %s = %s", enrolled.String(), "ENROLLED")
	}
	if sentCode.String() != "SENT_CODE" {
		t.Fatalf("Mismatch: %s = %s", sentCode.String(), "SENT_CODE")
	}
	if valid.String() != "VALID" {
		t.Fatalf("Mismatch: %s = %s", valid.String(), "VALID")
	}
}

func TestParseOTPStatusEnum(t *testing.T) {
	undef := ParseOTPStatusEnum("UNDEFINED")
	initial := ParseOTPStatusEnum("INITIAL")
	enrolled := ParseOTPStatusEnum("ENROLLED")
	sentCode := ParseOTPStatusEnum("SENT_CODE")
	valid := ParseOTPStatusEnum("VALID")

	if undef != OTP_STATUS_UNDEFINED {
		t.Fatalf("Mismatch: %s = %s", undef, OTP_STATUS_UNDEFINED)
	}
	if initial != OTP_STATUS_INITIAL {
		t.Fatalf("Mismatch: %s = %s", initial, OTP_STATUS_INITIAL)
	}
	if enrolled != OTP_STATUS_ENROLLED {
		t.Fatalf("Mismatch: %s = %s", enrolled, OTP_STATUS_ENROLLED)
	}
	if sentCode != OTP_STATUS_SENT_CODE {
		t.Fatalf("Mismatch: %s = %s", sentCode, OTP_STATUS_SENT_CODE)
	}
	if valid != OTP_STATUS_VALID {
		t.Fatalf("Mismatch: %s = %s", valid, OTP_STATUS_VALID)
	}
}

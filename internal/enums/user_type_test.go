package enums

import (
	pb "cchoice/proto"
	"testing"
)

func TestUserTypeToString(t *testing.T) {
	undef := USER_TYPE_UNDEFINED
	api := USER_TYPE_API
	system := USER_TYPE_SYSTEM

	if undef.String() != "UNDEFINED" {
		t.Fatalf("Mismatch: %s = %s", undef.String(), "UNDEFINED")
	}

	if api.String() != "API" {
		t.Fatalf("Mismatch: %s = %s", api.String(), "API")
	}

	if system.String() != "SYSTEM" {
		t.Fatalf("Mismatch: %s = %s", system.String(), "SYSTEM")
	}
}

func TestParseUserTypeEnum(t *testing.T) {
	undef := ParseUserTypeEnum("UNDEFINED")
	api := ParseUserTypeEnum("API")
	system := ParseUserTypeEnum("SYSTEM")

	if undef != USER_TYPE_UNDEFINED {
		t.Fatalf("Mismatch: %s = %s", undef, USER_TYPE_UNDEFINED)
	}
	if api != USER_TYPE_API {
		t.Fatalf("Mismatch: %s = %s", api, USER_TYPE_API)
	}
	if system != USER_TYPE_SYSTEM {
		t.Fatalf("Mismatch: %s = %s", system, USER_TYPE_SYSTEM)
	}
}

func TestParseUserTypeEnumPB(t *testing.T) {
	undef := ParseUserTypeEnumPB("UNDEFINED")
	api := ParseUserTypeEnumPB("API")
	system := ParseUserTypeEnumPB("SYSTEM")

	if undef != pb.UserType_UNDEFINED {
		t.Fatalf("Mismatch: %s = %s", undef, pb.UserType_UNDEFINED)
	}
	if api != pb.UserType_API {
		t.Fatalf("Mismatch: %s = %s", api, pb.UserType_API)
	}
	if system != pb.UserType_SYSTEM {
		t.Fatalf("Mismatch: %s = %s", system, pb.UserType_SYSTEM)
	}
}

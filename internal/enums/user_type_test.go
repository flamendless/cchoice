package enums

import (
	pb "cchoice/proto"
	"testing"
)

var tblUserType = map[UserType]string{
	USER_TYPE_UNDEFINED: "UNDEFINED",
	USER_TYPE_API:       "API",
	USER_TYPE_SYSTEM:    "SYSTEM",
}

func TestUserTypeToString(t *testing.T) {
	for usertype, val := range tblUserType {
		if usertype.String() != val {
			t.Fatalf("Mismatch: %s = %s", usertype.String(), val)
		}
	}
}

func TestParseUserTypeEnum(t *testing.T) {
	for usertype, val := range tblUserType {
		parsed := ParseUserTypeEnum(val)
		if parsed != usertype {
			t.Fatalf("Mismatch: %s = %s", val, usertype)
		}
	}
}

func TestParseUserTypeEnumPB(t *testing.T) {
	tbl := map[string]pb.UserType_UserType{
		"UNDEFINED": pb.UserType_UNDEFINED,
		"API":       pb.UserType_API,
		"SYSTEM":    pb.UserType_SYSTEM,
	}
	for val, usertype := range tbl {
		enum := StringToPBEnum(val, pb.UserType_UserType_value, pb.UserType_UNDEFINED)
		if enum != usertype {
			t.Fatalf("Mismatch: %s = %s", enum, usertype)
		}
	}
}

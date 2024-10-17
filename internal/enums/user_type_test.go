package enums

import (
	pb "cchoice/proto"
	"testing"

	"github.com/stretchr/testify/require"
)

var tblUserType = map[UserType]string{
	USER_TYPE_UNDEFINED: "UNDEFINED",
	USER_TYPE_API:       "API",
	USER_TYPE_SYSTEM:    "SYSTEM",
}

func TestUserTypeToString(t *testing.T) {
	for usertype, val := range tblUserType {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, val, usertype.String())
		})
	}
}

func TestParseUserTypeEnum(t *testing.T) {
	for usertype, val := range tblUserType {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, usertype, ParseUserTypeEnum(val))
		})
	}
}

func TestParseUserTypeEnumPB(t *testing.T) {
	tbl := map[string]pb.UserType_UserType{
		"UNDEFINED": pb.UserType_UNDEFINED,
		"API":       pb.UserType_API,
		"SYSTEM":    pb.UserType_SYSTEM,
	}
	for val, usertype := range tbl {
		t.Run(val, func(t *testing.T) {
			enum := StringToPBEnum(val, pb.UserType_UserType_value, pb.UserType_UNDEFINED)
			require.Equal(t, usertype, enum)
		})
	}
}

func BenchmarkUserTypeToString(b *testing.B) {
	for usertype := range tblUserType {
		for i := 0; i < b.N; i++ {
			_ = usertype.String()
		}
	}
}

func BenchmarkParseUserTypeEnum(b *testing.B) {
	for _, val := range tblUserType {
		for i := 0; i < b.N; i++ {
			_ = ParseUserTypeEnum(val)
		}
	}
}

func BenchmarkParseUserTypeEnumPB(b *testing.B) {
	tbl := map[string]pb.UserType_UserType{
		"UNDEFINED": pb.UserType_UNDEFINED,
		"API":       pb.UserType_API,
		"SYSTEM":    pb.UserType_SYSTEM,
	}
	for val := range tbl {
		for i := 0; i < b.N; i++ {
			_ = StringToPBEnum(val, pb.UserType_UserType_value, pb.UserType_UNDEFINED)
		}
	}
}

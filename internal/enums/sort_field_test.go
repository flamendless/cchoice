package enums

import (
	pb "cchoice/proto"
	"testing"

	"github.com/stretchr/testify/require"
)

var tblSortField = map[SortField]string{
	SORT_FIELD_UNDEFINED:  "UNDEFINED",
	SORT_FIELD_NAME:       "NAME",
	SORT_FIELD_CREATED_AT: "CREATED_AT",
}

func TestSortFieldToString(t *testing.T) {
	for sortfield, val := range tblSortField {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, val, sortfield.String())
		})
	}
}

func TestParseSortFieldEnum(t *testing.T) {
	for sortfield, val := range tblSortField {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, sortfield, ParseSortFieldEnum(val))
		})
	}
}

func TestParseSortFieldEnumPB(t *testing.T) {
	tbl := map[string]pb.SortField_SortField{
		"UNDEFINED":  pb.SortField_UNDEFINED,
		"NAME":       pb.SortField_NAME,
		"CREATED_AT": pb.SortField_CREATED_AT,
	}
	for val, sortfield := range tbl {
		t.Run(val, func(t *testing.T) {
			enum := StringToPBEnum(val, pb.SortField_SortField_value, pb.SortField_UNDEFINED)
			require.Equal(t, sortfield, enum)
		})
	}
}

func BenchmarkSortFieldToString(b *testing.B) {
	for sortfield := range tblSortField {
		for i := 0; i < b.N; i++ {
			_ = sortfield.String()
		}
	}
}

func BenchmarkParseSortFieldEnum(b *testing.B) {
	for _, val := range tblSortField {
		for i := 0; i < b.N; i++ {
			_ = ParseSortFieldEnum(val)
		}
	}
}

func BenchmarkParseSortFieldEnumPB(b *testing.B) {
	tbl := map[string]pb.SortField_SortField{
		"UNDEFINED":  pb.SortField_UNDEFINED,
		"NAME":       pb.SortField_NAME,
		"CREATED_AT": pb.SortField_CREATED_AT,
	}
	for val := range tbl {
		for i := 0; i < b.N; i++ {
			_ = StringToPBEnum(val, pb.SortField_SortField_value, pb.SortField_UNDEFINED)
		}
	}
}

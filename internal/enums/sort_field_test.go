package enums

import (
	pb "cchoice/proto"
	"testing"
)

var tblSortField = map[SortField]string{
	SORT_FIELD_UNDEFINED:  "UNDEFINED",
	SORT_FIELD_NAME:       "NAME",
	SORT_FIELD_CREATED_AT: "CREATED_AT",
}

func TestSortFieldToString(t *testing.T) {
	for sortfield, val := range tblSortField {
		if sortfield.String() != val {
			t.Fatalf("Mismatch: %s = %s", sortfield.String(), val)
		}
	}
}

func TestParseSortFieldEnum(t *testing.T) {
	for sortfield, val := range tblSortField {
		parsed := ParseSortFieldEnum(val)
		if parsed != sortfield {
			t.Fatalf("Mismatch: %s = %s", val, sortfield)
		}
	}
}

func TestParseSortFieldEnumPB(t *testing.T) {
	tbl := map[string]pb.SortField_SortField{
		"UNDEFINED":  pb.SortField_UNDEFINED,
		"NAME":       pb.SortField_NAME,
		"CREATED_AT": pb.SortField_CREATED_AT,
	}
	for val, sortfield := range tbl {
		enum := StringToPBEnum(val, pb.SortField_SortField_value, pb.SortField_UNDEFINED)
		if enum != sortfield {
			t.Fatalf("Mismatch: %s = %s", enum, sortfield)
		}
	}
}

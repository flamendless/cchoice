package enums

import (
	pb "cchoice/proto"
	"testing"
)

func TestSortFieldToString(t *testing.T) {
	undef := SORT_FIELD_UNDEFINED
	name := SORT_FIELD_NAME
	createdAt := SORT_FIELD_CREATED_AT

	if undef.String() != "UNDEFINED" {
		t.Fatalf("Mismatch: %s = %s", undef.String(), "UNDEFINED")
	}

	if name.String() != "NAME" {
		t.Fatalf("Mismatch: %s = %s", name.String(), "NAME")
	}

	if createdAt.String() != "CREATED_AT" {
		t.Fatalf("Mismatch: %s = %s", createdAt.String(), "CREATED_AT")
	}
}

func TestParseSortFieldEnum(t *testing.T) {
	undef := ParseSortFieldEnum("UNDEFINED")
	name := ParseSortFieldEnum("NAME")
	createdAt := ParseSortFieldEnum("CREATED_AT")

	if undef != SORT_FIELD_UNDEFINED {
		t.Fatalf("Mismatch: %s = %s", undef, SORT_FIELD_UNDEFINED)
	}
	if name != SORT_FIELD_NAME {
		t.Fatalf("Mismatch: %s = %s", name, SORT_FIELD_NAME)
	}
	if createdAt != SORT_FIELD_CREATED_AT {
		t.Fatalf("Mismatch: %s = %s", createdAt, SORT_FIELD_CREATED_AT)
	}
}

func TestParseSortFieldEnumPB(t *testing.T) {
	undef := ParseSortFieldEnumPB("UNDEFINED")
	name := ParseSortFieldEnumPB("NAME")
	createdAt := ParseSortFieldEnumPB("CREATED_AT")

	if undef != pb.SortField_UNDEFINED {
		t.Fatalf("Mismatch: %s = %s", undef, pb.SortField_UNDEFINED)
	}
	if name != pb.SortField_NAME {
		t.Fatalf("Mismatch: %s = %s", name, pb.SortField_NAME)
	}
	if createdAt != pb.SortField_CREATED_AT {
		t.Fatalf("Mismatch: %s = %s", createdAt, pb.SortField_CREATED_AT)
	}
}

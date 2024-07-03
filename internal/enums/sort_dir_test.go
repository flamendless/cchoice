package enums

import (
	pb "cchoice/proto"
	"testing"
)

func TestSortDirToString(t *testing.T) {
	undef := SORT_DIR_UNDEFINED
	asc := SORT_DIR_ASC
	desc := SORT_DIR_DESC

	if undef.String() != "UNDEFINED" {
		t.Fatalf("Mismatch: %s = %s", undef.String(), "UNDEFINED")
	}

	if asc.String() != "ASC" {
		t.Fatalf("Mismatch: %s = %s", asc.String(), "ASC")
	}

	if desc.String() != "DESC" {
		t.Fatalf("Mismatch: %s = %s", desc.String(), "DESC")
	}
}

func TestParseSortDirEnum(t *testing.T) {
	undef := ParseSortDirEnum("UNDEFINED")
	asc := ParseSortDirEnum("ASC")
	desc := ParseSortDirEnum("DESC")

	if undef != SORT_DIR_UNDEFINED {
		t.Fatalf("Mismatch: %s = %s", undef, SORT_DIR_UNDEFINED)
	}
	if asc != SORT_DIR_ASC {
		t.Fatalf("Mismatch: %s = %s", asc, SORT_DIR_ASC)
	}
	if desc != SORT_DIR_DESC {
		t.Fatalf("Mismatch: %s = %s", desc, SORT_DIR_DESC)
	}
}

func TestParseSortDirEnumPB(t *testing.T) {
	undef := ParseSortDirEnumPB("UNDEFINED")
	asc := ParseSortDirEnumPB("ASC")
	desc := ParseSortDirEnumPB("DESC")

	if undef != pb.SortDir_UNDEFINED {
		t.Fatalf("Mismatch: %s = %s", undef, pb.SortDir_UNDEFINED)
	}
	if asc != pb.SortDir_ASC {
		t.Fatalf("Mismatch: %s = %s", asc, pb.SortDir_ASC)
	}
	if desc != pb.SortDir_DESC {
		t.Fatalf("Mismatch: %s = %s", desc, pb.SortDir_DESC)
	}
}

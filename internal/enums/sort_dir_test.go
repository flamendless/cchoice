package enums

import (
	pb "cchoice/proto"
	"testing"
)

var tblSortDir = map[SortDir]string{
	SORT_DIR_UNDEFINED: "UNDEFINED",
	SORT_DIR_ASC:       "ASC",
	SORT_DIR_DESC:      "DESC",
}

func TestSortDirToString(t *testing.T) {
	for sortdir, val := range tblSortDir {
		if sortdir.String() != val {
			t.Fatalf("Mismatch: %s = %s", sortdir.String(), val)
		}
	}
}

func TestParseSortDirEnum(t *testing.T) {
	for sortdir, val := range tblSortDir {
		parsed := ParseSortDirEnum(val)
		if parsed != sortdir {
			t.Fatalf("Mismatch: %s = %s", val, sortdir)
		}
	}
}

func TestParseSortDirEnumPB(t *testing.T) {
	tbl := map[string]pb.SortDir_SortDir{
		"UNDEFINED": pb.SortDir_UNDEFINED,
		"ASC":       pb.SortDir_ASC,
		"DESC":      pb.SortDir_DESC,
	}
	for val, sortdir := range tbl {
		enum := StringToPBEnum(val, pb.SortDir_SortDir_value, pb.SortDir_UNDEFINED)
		if enum != sortdir {
			t.Fatalf("Mismatch: %s = %s", enum, sortdir)
		}
	}
}

func BenchmarkSortDirToString(b *testing.B) {
	for sortdir := range tblSortDir {
		for i := 0; i < b.N; i++ {
			_ = sortdir.String()
		}
	}
}

func BenchmarkParseSortDirEnum(b *testing.B) {
	for _, val := range tblSortDir {
		for i := 0; i < b.N; i++ {
			_ = ParseSortDirEnum(val)
		}
	}
}

func BenchmarkParseSortDirEnumPB(b *testing.B) {
	tbl := map[string]pb.SortDir_SortDir{
		"UNDEFINED": pb.SortDir_UNDEFINED,
		"ASC":       pb.SortDir_ASC,
		"DESC":      pb.SortDir_DESC,
	}
	for val := range tbl {
		for i := 0; i < b.N; i++ {
			_ = StringToPBEnum(val, pb.SortDir_SortDir_value, pb.SortDir_UNDEFINED)
		}
	}
}

package enums

import (
	pb "cchoice/proto"
	"testing"

	"github.com/stretchr/testify/require"
)

var tblSortDir = map[SortDir]string{
	SORT_DIR_UNDEFINED: "UNDEFINED",
	SORT_DIR_ASC:       "ASC",
	SORT_DIR_DESC:      "DESC",
}

func TestSortDirToString(t *testing.T) {
	for sortdir, val := range tblSortDir {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, val, sortdir.String())
		})
	}
}

func TestParseSortDirEnum(t *testing.T) {
	for sortdir, val := range tblSortDir {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, sortdir, ParseSortDirEnum(val))
		})
	}
}

func TestParseSortDirEnumPB(t *testing.T) {
	tbl := map[string]pb.SortDir_SortDir{
		"UNDEFINED": pb.SortDir_UNDEFINED,
		"ASC":       pb.SortDir_ASC,
		"DESC":      pb.SortDir_DESC,
	}
	for val, sortdir := range tbl {
		t.Run(val, func(t *testing.T) {
			enum := StringToPBEnum(val, pb.SortDir_SortDir_value, pb.SortDir_UNDEFINED)
			require.Equal(t, sortdir, enum)
		})
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

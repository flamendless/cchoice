package enums

import (
	pb "cchoice/proto"
	"fmt"
)

type SortDir int

const (
	UndefinedSortDir SortDir = iota
	ASC
	DESC
)

func (t SortDir) String() string {
	switch t {
	case UndefinedSortDir:
		return "UNDEFINED"
	case ASC:
		return "ASC"
	case DESC:
		return "DESC"
	default:
		panic("unknown enum")
	}
}

func ParseSortDirEnum(e string) SortDir {
	switch e {
	case "UNDEFINED":
		return UndefinedSortDir
	case "ASC":
		return ASC
	case "DESC":
		return DESC
	default:
		panic(fmt.Sprintf("Can't convert '%s' to SortDir enum", e))
	}
}

func ParseSortDirEnumPB(e string) pb.SortDir {
	switch e {
	case "ASC":
		return pb.SortDir_ASC
	case "DESC":
		return pb.SortDir_DESC
	default:
		return pb.SortDir_SORT_DIR_UNDEFINED
	}
}

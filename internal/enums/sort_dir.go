package enums

import (
	pb "cchoice/proto"
	"fmt"
)

type SortDir int

const (
	SORT_DIR_UNDEFINED SortDir = iota
	SORT_DIR_ASC
	SORT_DIR_DESC
)

func (t SortDir) String() string {
	switch t {
	case SORT_DIR_UNDEFINED:
		return "UNDEFINED"
	case SORT_DIR_ASC:
		return "ASC"
	case SORT_DIR_DESC:
		return "DESC"
	default:
		panic("unknown enum")
	}
}

func ParseSortDirEnum(e string) SortDir {
	switch e {
	case "UNDEFINED":
		return SORT_DIR_UNDEFINED
	case "ASC":
		return SORT_DIR_ASC
	case "DESC":
		return SORT_DIR_DESC
	default:
		panic(fmt.Sprintf("Can't convert '%s' to SortDir enum", e))
	}
}

func ParseSortDirEnumPB(e string) pb.SortDir_SortDir {
	switch e {
	case "ASC":
		return pb.SortDir_ASC
	case "DESC":
		return pb.SortDir_DESC
	default:
		return pb.SortDir_UNDEFINED
	}
}

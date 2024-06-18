package enums

import (
	pb "cchoice/proto"
	"fmt"
)

type SortField int

const (
	UndefinedSortField SortField = iota
	Name
)

func (t SortField) String() string {
	switch t {
	case UndefinedSortField:
		return "UNDEFINED"
	case Name:
		return "name"
	default:
		panic("unknown enum")
	}
}

func ParseSortFieldEnum(e string) SortField {
	switch e {
	case "UNDEFINED":
		return UndefinedSortField
	case "name":
		return Name
	default:
		panic(fmt.Sprintf("Can't convert '%s' to SortField enum", e))
	}
}

func ParseSortFieldEnumPB(e string) pb.SortField {
	switch e {
	case "UndefinedSortField":
		return *pb.SortField_SORT_FIELD_UNDEFINED.Enum()
	case "name":
		return pb.SortField_NAME
	default:
		panic(fmt.Sprintf("Can't convert '%s' to pb.SortField enum", e))
	}
}

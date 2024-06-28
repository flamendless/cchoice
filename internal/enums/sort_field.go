package enums

import (
	pb "cchoice/proto"
	"fmt"
)

type SortField int

const (
	UndefinedSortField SortField = iota
	Name
	CreatedAt
)

func (t SortField) String() string {
	switch t {
	case UndefinedSortField:
		return "UNDEFINED"
	case Name:
		return "name"
	case CreatedAt:
		return "created_t"
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
	case "created_at":
		return CreatedAt
	default:
		panic(fmt.Sprintf("Can't convert '%s' to SortField enum", e))
	}
}

func ParseSortFieldEnumPB(e string) pb.SortField {
	switch e {
	case "name":
		return pb.SortField_NAME
	case "created_at":
		return pb.SortField_CREATED_AT
	default:
		return pb.SortField_SORT_FIELD_UNDEFINED
	}
}

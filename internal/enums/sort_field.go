package enums

import (
	pb "cchoice/proto"
	"fmt"
)

type SortField int

const (
	SORT_FIELD_UNDEFINED SortField = iota
	SORT_FIELD_NAME
	SORT_FIELD_CREATED_AT
)

func (t SortField) String() string {
	switch t {
	case SORT_FIELD_UNDEFINED:
		return "UNDEFINED"
	case SORT_FIELD_NAME:
		return "name"
	case SORT_FIELD_CREATED_AT:
		return "created_t"
	default:
		panic("unknown enum")
	}
}

func ParseSortFieldEnum(e string) SortField {
	switch e {
	case "UNDEFINED":
		return SORT_FIELD_UNDEFINED
	case "name":
		return SORT_FIELD_NAME
	case "created_at":
		return SORT_FIELD_CREATED_AT
	default:
		panic(fmt.Sprintf("Can't convert '%s' to SortField enum", e))
	}
}

func ParseSortFieldEnumPB(e string) pb.SortField_SortField {
	switch e {
	case "name":
		return pb.SortField_NAME
	case "created_at":
		return pb.SortField_CREATED_AT
	default:
		return pb.SortField_UNDEFINED
	}
}

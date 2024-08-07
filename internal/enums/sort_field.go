package enums

type SortField int

const (
	SORT_FIELD_UNDEFINED SortField = iota
	SORT_FIELD_NAME
	SORT_FIELD_CREATED_AT
)

func (t SortField) String() string {
	switch t {
	case SORT_FIELD_NAME:
		return "NAME"
	case SORT_FIELD_CREATED_AT:
		return "CREATED_AT"
	default:
		return "UNDEFINED"
	}
}

func ParseSortFieldEnum(e string) SortField {
	switch e {
	case "NAME":
		return SORT_FIELD_NAME
	case "CREATED_AT":
		return SORT_FIELD_CREATED_AT
	default:
		return SORT_FIELD_UNDEFINED
	}
}

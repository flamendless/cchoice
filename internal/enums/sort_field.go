package enums

//go:generate stringer -type=SortField -trimprefix=SORT_FIELD_

type SortField int

const (
	SORT_FIELD_UNDEFINED SortField = iota
	SORT_FIELD_NAME
	SORT_FIELD_CREATED_AT
)

func ParseSortFieldEnum(e string) SortField {
	switch e {
	case SORT_FIELD_NAME.String():
		return SORT_FIELD_NAME
	case SORT_FIELD_CREATED_AT.String():
		return SORT_FIELD_CREATED_AT
	default:
		return SORT_FIELD_UNDEFINED
	}
}

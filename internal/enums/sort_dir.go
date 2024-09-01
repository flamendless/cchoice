package enums

//go:generate stringer -type=SortDir -trimprefix=SORT_DIR_

type SortDir int

const (
	SORT_DIR_UNDEFINED SortDir = iota
	SORT_DIR_ASC
	SORT_DIR_DESC
)

func ParseSortDirEnum(e string) SortDir {
	switch e {
	case SORT_DIR_ASC.String():
		return SORT_DIR_ASC
	case SORT_DIR_DESC.String():
		return SORT_DIR_DESC
	default:
		return SORT_DIR_UNDEFINED
	}
}

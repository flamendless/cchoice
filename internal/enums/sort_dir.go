package enums

type SortDir int

const (
	SORT_DIR_UNDEFINED SortDir = iota
	SORT_DIR_ASC
	SORT_DIR_DESC
)

func (t SortDir) String() string {
	switch t {
	case SORT_DIR_ASC:
		return "ASC"
	case SORT_DIR_DESC:
		return "DESC"
	default:
		return "UNDEFINED"
	}
}

func ParseSortDirEnum(e string) SortDir {
	switch e {
	case "ASC":
		return SORT_DIR_ASC
	case "DESC":
		return SORT_DIR_DESC
	default:
		return SORT_DIR_UNDEFINED
	}
}

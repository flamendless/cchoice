package database

//go:generate go tool stringer -type=DBPrefix -trimprefix=DB_PREFIX_

type DBPrefix int

const (
	DB_PREFIX_UNDEFINED DBPrefix = iota
	DB_PREFIX_CATEGORY
)

func ParseDBPrefixToEnum(dbprefix string) DBPrefix {
	switch dbprefix {
	case DB_PREFIX_CATEGORY.String():
		return DB_PREFIX_CATEGORY
	default:
		return DB_PREFIX_UNDEFINED
	}
}

package enums

import "strings"

//go:generate go tool stringer -type=ListingSortDirection -trimprefix=LISTING_SORT_DIRECTION_

type ListingSortDirection int

const (
	LISTING_SORT_DIRECTION_UNDEFINED ListingSortDirection = iota
	LISTING_SORT_DIRECTION_ASC
	LISTING_SORT_DIRECTION_DESC
)

func ParseListingSortDirection(s string) ListingSortDirection {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case LISTING_SORT_DIRECTION_ASC.String():
		return LISTING_SORT_DIRECTION_ASC
	case LISTING_SORT_DIRECTION_DESC.String():
		return LISTING_SORT_DIRECTION_DESC
	default:
		return LISTING_SORT_DIRECTION_UNDEFINED
	}
}

func (d ListingSortDirection) IsAscending() bool {
	return d == LISTING_SORT_DIRECTION_ASC
}

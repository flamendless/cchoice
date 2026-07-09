package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=ThemeStatus -trimprefix=THEME_STATUS_

type ThemeStatus int

const (
	THEME_STATUS_UNDEFINED ThemeStatus = iota
	THEME_STATUS_DRAFT
	THEME_STATUS_PUBLISHED
	THEME_STATUS_DELETED
)

var AllThemeStatuses = []ThemeStatus{
	THEME_STATUS_DRAFT,
	THEME_STATUS_PUBLISHED,
	THEME_STATUS_DELETED,
}

func ParseThemeStatusToEnum(ts string) ThemeStatus {
	switch strings.ToUpper(ts) {
	case THEME_STATUS_DRAFT.String():
		return THEME_STATUS_DRAFT
	case THEME_STATUS_PUBLISHED.String():
		return THEME_STATUS_PUBLISHED
	case THEME_STATUS_DELETED.String():
		return THEME_STATUS_DELETED
	default:
		return THEME_STATUS_UNDEFINED
	}
}

func MustParseThemeStatusToEnum(ts string) ThemeStatus {
	res := ParseThemeStatusToEnum(ts)
	if res == THEME_STATUS_UNDEFINED {
		panic(fmt.Sprintf("Unexpected ThemeStatus. Got '%s'", ts))
	}
	return res
}

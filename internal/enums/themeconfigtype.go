package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=ThemeConfigType -trimprefix=THEME_CONFIG_TYPE_

type ThemeConfigType int

const (
	THEME_CONFIG_TYPE_UNDEFINED ThemeConfigType = iota
	THEME_CONFIG_TYPE_JSON
	THEME_CONFIG_TYPE_INI
)

var AllThemeConfigTypes = []ThemeConfigType{
	THEME_CONFIG_TYPE_JSON,
	THEME_CONFIG_TYPE_INI,
}

func ParseThemeConfigTypeToEnum(t string) ThemeConfigType {
	switch strings.ToUpper(t) {
	case THEME_CONFIG_TYPE_JSON.String():
		return THEME_CONFIG_TYPE_JSON
	case THEME_CONFIG_TYPE_INI.String():
		return THEME_CONFIG_TYPE_INI
	default:
		return THEME_CONFIG_TYPE_UNDEFINED
	}
}

func MustParseThemeConfigTypeToEnum(t string) ThemeConfigType {
	res := ParseThemeConfigTypeToEnum(t)
	if res == THEME_CONFIG_TYPE_UNDEFINED {
		panic(fmt.Sprintf("Unexpected ThemeConfigType. Got '%s'", t))
	}
	return res
}

package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=TimeOff -trimprefix=TIME_OFF_

type TimeOff int

const (
	TIME_OFF_UNDEFINED TimeOff = iota
	TIME_OFF_VL
	TIME_OFF_SL
	TIME_OFF_ABSENT
)

func ParseTimeOffToEnum(to string) TimeOff {
	switch strings.ToUpper(to) {
	case "VL":
		return TIME_OFF_VL
	case "SL":
		return TIME_OFF_SL
	case "ABSENT":
		return TIME_OFF_ABSENT
	default:
		return TIME_OFF_UNDEFINED
	}
}

func MustParseTimeOffToEnum(to string) TimeOff {
	res := ParseTimeOffToEnum(to)
	if res == TIME_OFF_UNDEFINED {
		panic(fmt.Sprintf("Unexpected type. Got '%s'", to))
	}
	return res
}

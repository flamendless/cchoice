package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=HolidayType -trimprefix=HOLIDAY_TYPE_

type HolidayType int

const (
	HOLIDAY_TYPE_UNDEFINED HolidayType = iota
	HOLIDAY_TYPE_PUBLIC
	HOLIDAY_TYPE_REGULAR
	HOLIDAY_TYPE_SPECIAL_NON_WORKING
	HOLIDAY_TYPE_SPECIAL_WORKING
)

var AllHolidayTypes = []HolidayType{
	HOLIDAY_TYPE_PUBLIC,
	HOLIDAY_TYPE_REGULAR,
	HOLIDAY_TYPE_SPECIAL_NON_WORKING,
	HOLIDAY_TYPE_SPECIAL_WORKING,
}

func ParseHolidayTypeToEnum(ht string) HolidayType {
	switch strings.ToUpper(ht) {
	case HOLIDAY_TYPE_PUBLIC.String():
		return HOLIDAY_TYPE_PUBLIC
	case HOLIDAY_TYPE_REGULAR.String():
		return HOLIDAY_TYPE_REGULAR
	case HOLIDAY_TYPE_SPECIAL_NON_WORKING.String():
		return HOLIDAY_TYPE_SPECIAL_NON_WORKING
	case HOLIDAY_TYPE_SPECIAL_WORKING.String():
		return HOLIDAY_TYPE_SPECIAL_WORKING
	default:
		return HOLIDAY_TYPE_UNDEFINED
	}
}

func MustParseHolidayTypeToEnum(ht string) HolidayType {
	res := ParseHolidayTypeToEnum(ht)
	if res == HOLIDAY_TYPE_UNDEFINED {
		panic(fmt.Sprintf("Unexpected HolidayType. Got '%s'", ht))
	}
	return res
}

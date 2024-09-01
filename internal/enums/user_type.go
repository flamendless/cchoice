package enums

//go:generate stringer -type=UserType -trimprefix=USER_TYPE_

import (
	"fmt"
)

type UserType int

const (
	USER_TYPE_UNDEFINED UserType = iota
	USER_TYPE_API
	USER_TYPE_SYSTEM
)

func ParseUserTypeEnum(e string) UserType {
	switch e {
	case USER_TYPE_UNDEFINED.String():
		return USER_TYPE_UNDEFINED
	case USER_TYPE_API.String():
		return USER_TYPE_API
	case USER_TYPE_SYSTEM.String():
		return USER_TYPE_SYSTEM
	default:
		panic(fmt.Sprintf("Can't convert '%s' to UserType enum", e))
	}
}

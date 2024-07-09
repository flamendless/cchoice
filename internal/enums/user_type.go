package enums

import (
	"fmt"
)

type UserType int

const (
	USER_TYPE_UNDEFINED UserType = iota
	USER_TYPE_API
	USER_TYPE_SYSTEM
)

func (t UserType) String() string {
	switch t {
	case USER_TYPE_UNDEFINED:
		return "UNDEFINED"
	case USER_TYPE_API:
		return "API"
	case USER_TYPE_SYSTEM:
		return "SYSTEM"
	default:
		panic("unknown enum")
	}
}

func ParseUserTypeEnum(e string) UserType {
	switch e {
	case "UNDEFINED":
		return USER_TYPE_UNDEFINED
	case "API":
		return USER_TYPE_API
	case "SYSTEM":
		return USER_TYPE_SYSTEM
	default:
		panic(fmt.Sprintf("Can't convert '%s' to UserType enum", e))
	}
}

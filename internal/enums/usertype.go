package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=UserType -trimprefix=USER_TYPE_

type UserType int

const (
	USER_TYPE_UNDEFINED UserType = iota
	USER_TYPE_STAFF
	USER_TYPE_CUSTOMER
)

func ParseUserTypeToEnum(e string) UserType {
	switch strings.ToUpper(e) {
	case USER_TYPE_STAFF.String():
		return USER_TYPE_STAFF
	case USER_TYPE_CUSTOMER.String():
		return USER_TYPE_CUSTOMER
	default:
		return USER_TYPE_UNDEFINED
	}
}

func MustParseUserTypeToEnum(e string) UserType {
	switch strings.ToUpper(e) {
	case USER_TYPE_STAFF.String():
		return USER_TYPE_STAFF
	case USER_TYPE_CUSTOMER.String():
		return USER_TYPE_CUSTOMER
	default:
		panic(fmt.Sprintf("Invalid UserType. Got '%s'", e))
	}
}

func (ut UserType) IsValid() bool {
	switch ut {
	case USER_TYPE_STAFF, USER_TYPE_CUSTOMER:
		return true
	default:
		return false
	}
}

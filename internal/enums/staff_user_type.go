package enums

import "fmt"

//go:generate go tool stringer -type=StaffUserType -trimprefix=STAFF_USER_TYPE_

type StaffUserType int

const (
	STAFF_USER_TYPE_UNDEFINED StaffUserType = iota
	STAFF_USER_TYPE_STAFF
	STAFF_USER_TYPE_SUPERUSER
)

func ParseStaffUserTypeToEnum(e string) StaffUserType {
	switch e {
	case STAFF_USER_TYPE_STAFF.String():
		return STAFF_USER_TYPE_STAFF
	case STAFF_USER_TYPE_SUPERUSER.String():
		return STAFF_USER_TYPE_SUPERUSER
	default:
		return STAFF_USER_TYPE_UNDEFINED
	}
}

func MustParseStaffUserTypeToEnum(e string) StaffUserType {
	switch e {
	case STAFF_USER_TYPE_STAFF.String():
		return STAFF_USER_TYPE_STAFF
	case STAFF_USER_TYPE_SUPERUSER.String():
		return STAFF_USER_TYPE_SUPERUSER
	default:
		panic(fmt.Sprintf("Invalid StaffUserType. Got '%s'", e))
	}
}

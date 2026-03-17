package enums

//go:generate go tool stringer -type=StaffRole -trimprefix=STAFF_ROLE_

type StaffRole int

const (
	STAFF_ROLE_UNDEFINED StaffRole = iota
	STAFF_ROLE_CREATE_PRODUCT
)

func ParseStaffRoleToEnum(e string) StaffRole {
	switch e {
	case STAFF_ROLE_CREATE_PRODUCT.String():
		return STAFF_ROLE_CREATE_PRODUCT
	default:
		return STAFF_ROLE_UNDEFINED
	}
}

func MustParseStaffRoleToEnum(e string) StaffRole {
	switch e {
	case STAFF_ROLE_CREATE_PRODUCT.String():
		return STAFF_ROLE_CREATE_PRODUCT
	default:
		panic("Invalid StaffRole. Got '" + e + "'")
	}
}

func (r StaffRole) IsValid() bool {
	return r != STAFF_ROLE_UNDEFINED
}

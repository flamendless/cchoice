package enums

//go:generate go tool stringer -type=StaffRole -trimprefix=STAFF_ROLE_

type StaffRole int

const (
	STAFF_ROLE_UNDEFINED StaffRole = iota
	STAFF_ROLE_CREATE_PRODUCT
	STAFF_ROLE_CREATE_CPOINTS
	STAFF_ROLE_MANAGE_HOLIDAYS
)

func ParseStaffRoleToEnum(e string) StaffRole {
	switch e {
	case STAFF_ROLE_CREATE_PRODUCT.String():
		return STAFF_ROLE_CREATE_PRODUCT
	case STAFF_ROLE_CREATE_CPOINTS.String():
		return STAFF_ROLE_CREATE_CPOINTS
	case STAFF_ROLE_MANAGE_HOLIDAYS.String():
		return STAFF_ROLE_MANAGE_HOLIDAYS
	default:
		return STAFF_ROLE_UNDEFINED
	}
}

func MustParseStaffRoleToEnum(e string) StaffRole {
	switch e {
	case STAFF_ROLE_CREATE_PRODUCT.String():
		return STAFF_ROLE_CREATE_PRODUCT
	case STAFF_ROLE_CREATE_CPOINTS.String():
		return STAFF_ROLE_CREATE_CPOINTS
	case STAFF_ROLE_MANAGE_HOLIDAYS.String():
		return STAFF_ROLE_MANAGE_HOLIDAYS
	default:
		panic("Invalid StaffRole. Got '" + e + "'")
	}
}

func (r StaffRole) IsValid() bool {
	return r != STAFF_ROLE_UNDEFINED
}

func GetAllStaffRoles() []StaffRole {
	return []StaffRole{
		STAFF_ROLE_CREATE_PRODUCT,
		STAFF_ROLE_CREATE_CPOINTS,
		STAFF_ROLE_MANAGE_HOLIDAYS,
	}
}

func RoleExists(role StaffRole) bool {
	return role.IsValid()
}

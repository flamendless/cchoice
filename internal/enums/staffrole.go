package enums

//go:generate go tool stringer -type=StaffRole -trimprefix=STAFF_ROLE_

type StaffRole int

const (
	STAFF_ROLE_UNDEFINED StaffRole = iota
	STAFF_ROLE_CREATE_PRODUCT
	STAFF_ROLE_CREATE_CPOINTS
	STAFF_ROLE_MANAGE_HOLIDAYS
	STAFF_ROLE_MANAGE_BRANDS
	STAFF_ROLE_MANAGE_PROMOS
	STAFF_ROLE_MANAGE_TRACKED_LINKS
)

func ParseStaffRoleToEnum(e string) StaffRole {
	switch e {
	case STAFF_ROLE_CREATE_PRODUCT.String():
		return STAFF_ROLE_CREATE_PRODUCT
	case STAFF_ROLE_CREATE_CPOINTS.String():
		return STAFF_ROLE_CREATE_CPOINTS
	case STAFF_ROLE_MANAGE_HOLIDAYS.String():
		return STAFF_ROLE_MANAGE_HOLIDAYS
	case STAFF_ROLE_MANAGE_BRANDS.String():
		return STAFF_ROLE_MANAGE_BRANDS
	case STAFF_ROLE_MANAGE_PROMOS.String():
		return STAFF_ROLE_MANAGE_PROMOS
	case STAFF_ROLE_MANAGE_TRACKED_LINKS.String():
		return STAFF_ROLE_MANAGE_TRACKED_LINKS
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
	case STAFF_ROLE_MANAGE_BRANDS.String():
		return STAFF_ROLE_MANAGE_BRANDS
	case STAFF_ROLE_MANAGE_PROMOS.String():
		return STAFF_ROLE_MANAGE_PROMOS
	case STAFF_ROLE_MANAGE_TRACKED_LINKS.String():
		return STAFF_ROLE_MANAGE_TRACKED_LINKS
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
		STAFF_ROLE_MANAGE_BRANDS,
		STAFF_ROLE_MANAGE_PROMOS,
		STAFF_ROLE_MANAGE_TRACKED_LINKS,
	}
}

func RoleExists(role StaffRole) bool {
	return role.IsValid()
}

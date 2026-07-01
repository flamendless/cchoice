package components

import "cchoice/internal/enums"

func staffListRowClass(status enums.StaffStatus) string {
	switch status {
	case enums.STAFF_STATUS_PROBATION:
		return "bg-yellow-50"
	case enums.STAFF_STATUS_REGULAR:
		return "bg-green-50"
	case enums.STAFF_STATUS_RESIGNED:
		return "bg-red-50"
	case enums.STAFF_STATUS_PART_TIME:
		return "bg-sky-50"
	case enums.STAFF_STATUS_OJT:
		return "bg-orange-50"
	default:
		return ""
	}
}

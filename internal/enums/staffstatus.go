package enums

import "strings"

//go:generate go tool stringer -type=StaffStatus -trimprefix=STAFF_STATUS_

type StaffStatus int

const (
	STAFF_STATUS_UNDEFINED StaffStatus = iota
	STAFF_STATUS_PROBATION
	STAFF_STATUS_REGULAR
	STAFF_STATUS_RESIGNED
	STAFF_STATUS_PART_TIME
	STAFF_STATUS_OJT
)

var AllStaffStatuses = []StaffStatus{
	STAFF_STATUS_PROBATION,
	STAFF_STATUS_REGULAR,
	STAFF_STATUS_RESIGNED,
	STAFF_STATUS_PART_TIME,
	STAFF_STATUS_OJT,
}

func ParseStaffStatusToEnum(e string) StaffStatus {
	switch strings.ToUpper(e) {
	case STAFF_STATUS_PROBATION.String():
		return STAFF_STATUS_PROBATION
	case STAFF_STATUS_REGULAR.String():
		return STAFF_STATUS_REGULAR
	case STAFF_STATUS_RESIGNED.String():
		return STAFF_STATUS_RESIGNED
	case STAFF_STATUS_PART_TIME.String():
		return STAFF_STATUS_PART_TIME
	case STAFF_STATUS_OJT.String():
		return STAFF_STATUS_OJT
	default:
		return STAFF_STATUS_UNDEFINED
	}
}

func (s StaffStatus) IsValid() bool {
	return s != STAFF_STATUS_UNDEFINED
}

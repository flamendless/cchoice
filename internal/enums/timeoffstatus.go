package enums

import "strings"

//go:generate go tool stringer -type=TimeOffStatus -trimprefix=TIME_OFF_STATUS_

type TimeOffStatus int

const (
	TIME_OFF_STATUS_UNDEFINED TimeOffStatus = iota
	TIME_OFF_STATUS_PENDING
	TIME_OFF_STATUS_APPROVED
	TIME_OFF_STATUS_REJECTED
)

func ParseTimeOffStatusToEnum(s string) TimeOffStatus {
	switch strings.ToUpper(s) {
	case TIME_OFF_STATUS_PENDING.String():
		return TIME_OFF_STATUS_PENDING
	case TIME_OFF_STATUS_APPROVED.String():
		return TIME_OFF_STATUS_APPROVED
	case TIME_OFF_STATUS_REJECTED.String():
		return TIME_OFF_STATUS_REJECTED
	default:
		return TIME_OFF_STATUS_UNDEFINED
	}
}

func (s TimeOffStatus) ToHuman() string {
	switch s {
	case TIME_OFF_STATUS_PENDING:
		return "Pending"
	case TIME_OFF_STATUS_APPROVED:
		return "Approved"
	case TIME_OFF_STATUS_REJECTED:
		return "Rejected"
	default:
		return "Unknown"
	}
}

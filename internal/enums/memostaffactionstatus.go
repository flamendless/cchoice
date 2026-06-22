package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=MemoStaffActionStatus -trimprefix=MEMO_STAFF_ACTION_STATUS_

type MemoStaffActionStatus int

const (
	MEMO_STAFF_ACTION_STATUS_UNDEFINED MemoStaffActionStatus = iota
	MEMO_STAFF_ACTION_STATUS_ACCEPTED
	MEMO_STAFF_ACTION_STATUS_REJECTED
)

var AllMemoStaffActionStatuses = []MemoStaffActionStatus{
	MEMO_STAFF_ACTION_STATUS_ACCEPTED,
	MEMO_STAFF_ACTION_STATUS_REJECTED,
}

func ParseMemoStaffActionStatusToEnum(s string) MemoStaffActionStatus {
	switch strings.ToUpper(s) {
	case MEMO_STAFF_ACTION_STATUS_ACCEPTED.String():
		return MEMO_STAFF_ACTION_STATUS_ACCEPTED
	case MEMO_STAFF_ACTION_STATUS_REJECTED.String():
		return MEMO_STAFF_ACTION_STATUS_REJECTED
	default:
		return MEMO_STAFF_ACTION_STATUS_UNDEFINED
	}
}

func MustParseMemoStaffActionStatusToEnum(s string) MemoStaffActionStatus {
	res := ParseMemoStaffActionStatusToEnum(s)
	if res == MEMO_STAFF_ACTION_STATUS_UNDEFINED {
		panic(fmt.Sprintf("Unexpected MemoStaffActionStatus. Got '%s'", s))
	}
	return res
}

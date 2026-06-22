package components

import (
	"slices"

	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/utils"
)

func memoTodayISO() string {
	return utils.NowPH().Format(constants.DateLayoutISO)
}

func memoEndDateMin(startDate string) string {
	today := memoTodayISO()
	if startDate > today {
		return startDate
	}
	return today
}

func isMemoStaffSelected(selectedIDs []string, staffID string) bool {
	return slices.Contains(selectedIDs, staffID)
}

func memoStaffDisplayName(name, staffID, currentStaffID string) string {
	if staffID == currentStaffID {
		return name + " (self)"
	}
	return name
}

func memoRecipientRowClass(status enums.MemoStaffActionStatus) string {
	switch status {
	case enums.MEMO_STAFF_ACTION_STATUS_ACCEPTED:
		return "bg-green-50"
	case enums.MEMO_STAFF_ACTION_STATUS_REJECTED:
		return "bg-red-50"
	default:
		return "bg-gray-50"
	}
}

func memoRecipientStatusLabel(status enums.MemoStaffActionStatus) string {
	switch status {
	case enums.MEMO_STAFF_ACTION_STATUS_ACCEPTED:
		return "Accepted"
	case enums.MEMO_STAFF_ACTION_STATUS_REJECTED:
		return "Rejected"
	default:
		return "Pending"
	}
}

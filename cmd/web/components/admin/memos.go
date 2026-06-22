package components

import (
	"slices"
	"strings"
	"time"

	"cchoice/cmd/web/models"
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

func memoCanSendEmails(memo models.AdminMemoListItem, currentStaffID string, isSuperuser bool) bool {
	if memo.Status != enums.MEMO_STATUS_PUBLISHED {
		return false
	}
	return isSuperuser || memo.CreatedByID == currentStaffID
}

func memoEmailOnCooldown(emailsSentAt string) bool {
	if emailsSentAt == "" || strings.HasPrefix(emailsSentAt, "1970-01-01") {
		return false
	}
	sentAt, err := time.Parse(constants.DateTimeLayoutISO, emailsSentAt)
	if err != nil {
		sentAt, err = time.Parse(constants.DateTimeLayoutTZISO, emailsSentAt)
		if err != nil {
			return false
		}
	}
	return time.Since(sentAt) < 24*time.Hour
}

func memoSendEmailsTooltip(onCooldown bool) string {
	if onCooldown {
		return "Emails sent recently"
	}
	return "Send emails"
}

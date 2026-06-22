package services

import (
	"database/sql"
	"time"

	"cchoice/internal/enums"
)

type Memo struct {
	ID           int64
	Title        string
	Message      string
	FileURL      string
	Status       enums.MemoStatus
	StartDate    string
	EndDate      string
	CreatedBy    int64
	CreatedAt    time.Time
	UpdatedAt    sql.NullString
	DeletedAt    string
	EmailsSentAt string
}

type MemoListItem struct {
	Memo
	CreatedByName     string
	CreatorPosition   string
}

type MemoRecipientRow struct {
	StaffID      string
	StaffName    string
	Email        string
	Position     string
	UserType     enums.StaffUserType
	ActionStatus enums.MemoStaffActionStatus
	RejectReason string
	AcceptedAt   string
	RejectedAt   string
}

type StaffPendingMemo struct {
	ID      string
	Title   string
	Message string
	FileURL string
}

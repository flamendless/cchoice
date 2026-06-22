package services

import (
	"database/sql"
	"time"

	"cchoice/internal/enums"
)

type Memo struct {
	ID        int64
	Title     string
	Message   string
	FileURL   string
	Status    enums.MemoStatus
	StartDate string
	EndDate   string
	CreatedBy int64
	CreatedAt time.Time
	UpdatedAt sql.NullString
	DeletedAt string
}

type MemoListItem struct {
	Memo
	CreatedByName string
}

type MemoRecipientRow struct {
	StaffID      string
	StaffName    string
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

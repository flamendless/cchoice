package models

import (
	"cchoice/internal/enums"

	"github.com/a-h/templ"
)

type StaffCard struct {
	Link        string
	Title       string
	Description string
	Icon        templ.Component
}

type AdminPromoListItem struct {
	ID          string
	Title       string
	Description string
	MediaURL    string
	StartDate   string
	EndDate     string
	Type        enums.PromoType
	Status      enums.PromoStatus
	BannerOnly  bool
	Priority    int64
	CreatedAt   string
}

type AdminMemoListItem struct {
	ID            string
	Title         string
	Message       string
	FileURL       string
	Status        enums.MemoStatus
	StartDate     string
	EndDate       string
	CreatedByName string
	CreatedAt     string
	RecipientIDs  []string
}

type AdminMemoRecipientRow struct {
	StaffID      string
	StaffName    string
	ActionStatus enums.MemoStaffActionStatus
	RejectReason string
	AcceptedAt   string
	RejectedAt   string
}

type StaffMemoCard struct {
	ID      string
	Title   string
	Message string
	FileURL string
}

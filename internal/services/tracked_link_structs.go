package services

import (
	"database/sql"
	"time"

	"cchoice/internal/enums"
)

type TrackedLink struct {
	ID             string
	Name           string
	Slug           string
	DestinationURL string
	Source         enums.TrackedLinkSource
	Medium         enums.TrackedLinkMedium
	Campaign       sql.NullString
	Status         enums.TrackedLinkStatus
	StaffID        sql.NullString
	CreatedAt      string
	UpdatedAt      string
}

type LinkClick struct {
	ID          int64
	LinkID      string
	ClickedAt   time.Time
	Referrer    sql.NullString
	UserAgent   sql.NullString
	IPHash      sql.NullString
	Device      sql.NullString
	UTMSource   sql.NullString
	UTMMedium   sql.NullString
	UTMCampaign sql.NullString
}

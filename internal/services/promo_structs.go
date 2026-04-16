package services

import (
	"database/sql"
	"time"

	"cchoice/internal/enums"
)

type Promo struct {
	ID          int64
	Title       string
	Description string
	MediaURL    string
	StartDate   string
	EndDate     string
	Type        enums.PromoType
	Status      enums.PromoStatus
	CreatedAt   time.Time
	UpdatedAt   sql.NullString
	DeletedAt   string
}

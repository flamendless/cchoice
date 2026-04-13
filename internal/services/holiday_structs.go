package services

import (
	"database/sql"
	"time"

	"cchoice/internal/enums"
)

type Holiday struct {
	CreatedAt time.Time      `json:"created_at"`
	Date      string         `json:"date"`
	Name      string         `json:"name"`
	UpdatedAt sql.NullString `json:"updated_at"`
	ID        int64
	Type      enums.HolidayType `json:"type"`
}

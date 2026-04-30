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

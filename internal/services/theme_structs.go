package services

import (
	"time"

	"cchoice/internal/enums"
)

type Theme struct {
	ID                int64
	Title             string
	Status            enums.ThemeStatus
	StartDate         string
	EndDate           string
	Configuration     string
	ConfigurationType enums.ThemeConfigType
	Active            bool
	CreatedBy         int64
	CreatedAt         time.Time
	UpdatedAt         string
	DeletedAt         string
}

package services

import (
	"time"

	"cchoice/internal/enums"
)

type Brand struct {
	CreatedAt    time.Time
	Name         string
	LogoS3URL    string
	ID           int64
	BrandImageID int64
	ProductCount int64
	Status       enums.BrandStatus
}

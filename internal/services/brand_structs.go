package services

import "time"

type Brand struct {
	ID           int64
	Name         string
	LogoS3URL    string
	BrandImageID int64
	ProductCount int64
	CreatedAt    time.Time
}


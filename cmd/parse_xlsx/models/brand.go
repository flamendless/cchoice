package models

import (
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"context"
	"time"
)

type Brand struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	Name      string
	ID        int64
}

type BrandImage struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	Path      string
	ID        int64
	BrandID   int64
	IsMain    bool
}

func NewBrand(brandName string) *Brand {
	now := time.Now().UTC()
	return &Brand{
		Name:      brandName,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: constants.DT_BEGINNING,
	}
}

func (brand *Brand) GetDBID(db database.Service) int64 {
	ctx := context.Background()
	existingBrandID, err := db.GetQueries().GetBrandIDByName(ctx, brand.Name)
	if err != nil {
		return 0
	}
	brand.ID = existingBrandID
	return existingBrandID
}

func (brand *Brand) InsertToDB(db database.Service) (int64, error) {
	ctx := context.Background()
	newBrandID, err := db.GetQueries().CreateBrand(ctx, queries.CreateBrandParams{
		Name:      brand.Name,
		CreatedAt: brand.CreatedAt,
		UpdatedAt: brand.UpdatedAt,
		DeletedAt: brand.UpdatedAt,
	})
	if err != nil {
		return 0, err
	}
	brand.ID = newBrandID
	return newBrandID, nil
}

func (brandImage *BrandImage) InsertToDB(db database.Service) (int64, error) {
	ctx := context.Background()
	newBrandImageID, err := db.GetQueries().CreateBrandImage(
		ctx,
		queries.CreateBrandImageParams{
			BrandID:   brandImage.BrandID,
			Path:      brandImage.Path,
			IsMain:    brandImage.IsMain,
			CreatedAt: brandImage.CreatedAt,
			UpdatedAt: brandImage.UpdatedAt,
			DeletedAt: brandImage.DeletedAt,
		},
	)
	if err != nil {
		return 0, err
	}
	brandImage.ID = newBrandImageID
	return newBrandImageID, nil
}

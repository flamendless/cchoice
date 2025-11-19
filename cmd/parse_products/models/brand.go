package models

import (
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"context"
	"database/sql"
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
	S3URL     string
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
		DeletedAt: constants.DtBeginning,
	}
}

func (brand *Brand) GetDBID(ctx context.Context, db database.Service) int64 {
	existingBrandID, err := db.GetQueries().GetBrandsIDByName(ctx, brand.Name)
	if err != nil {
		return 0
	}
	brand.ID = existingBrandID
	return existingBrandID
}

func (brand *Brand) InsertToDB(ctx context.Context, db database.Service) (int64, error) {
	newBrandID, err := db.GetQueries().CreateBrands(ctx, queries.CreateBrandsParams{
		Name:      brand.Name,
		CreatedAt: brand.CreatedAt,
		UpdatedAt: brand.UpdatedAt,
		DeletedAt: brand.DeletedAt,
	})
	if err != nil {
		return 0, err
	}
	brand.ID = newBrandID
	return newBrandID, nil
}

func (brandImage *BrandImage) InsertToDB(ctx context.Context, db database.Service) (int64, error) {
	newBrandImageID, err := db.GetQueries().CreateBrandImages(
		ctx,
		queries.CreateBrandImagesParams{
			BrandID:   brandImage.BrandID,
			Path:      brandImage.Path,
			S3Url:     sql.NullString{Valid: brandImage.S3URL != "", String: brandImage.S3URL},
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

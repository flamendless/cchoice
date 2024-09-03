package models

import (
	"cchoice/cchoice_db"
	"cchoice/internal/constants"
	"cchoice/internal/ctx"
	"context"
	"time"
)

type Brand struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type BrandImage struct {
	ID        int64
	BrandID   int64
	Path      string
	IsMain    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
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

func (brand *Brand) GetDBID(ctxDB *ctx.Database) int64 {
	ctx := context.Background()
	existingBrandID, err := ctxDB.QueriesRead.GetBrandIDByName(ctx, brand.Name)
	if err != nil {
		return 0
	}
	brand.ID = existingBrandID
	return existingBrandID
}

func (brand *Brand) InsertToDB(ctxDB *ctx.Database) (int64, error) {
	ctx := context.Background()
	newBrandID, err := ctxDB.Queries.CreateBrand(ctx, cchoice_db.CreateBrandParams{
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

func (brandImage *BrandImage) InsertToDB(ctxDB *ctx.Database) (int64, error) {
	ctx := context.Background()
	newBrandImageID, err := ctxDB.Queries.CreateBrandImage(
		ctx,
		cchoice_db.CreateBrandImageParams{
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

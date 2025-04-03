package models

import (
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"context"
	"database/sql"
	"time"
)

type ProductImage struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	Product   *Product
	Path      string
	Thumbnail string
	ID        int64
}

func (pi *ProductImage) InsertToDB(ctx context.Context, db database.Service) (int64, error) {
	if pi == nil {
		panic("nil ProductImage")
	}
	if pi.Product == nil {
		panic("nil ProductImage.Product")
	}
	now := time.Now().UTC()
	insertedProductImage, err := db.GetQueries().CreateProductImage(ctx, queries.CreateProductImageParams{
		ProductID: pi.Product.ID,
		Path:      pi.Path,
		Thumbnail: sql.NullString{Valid: pi.Thumbnail != "", String: pi.Thumbnail},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: constants.DT_BEGINNING,
	})
	if err != nil {
		return 0, err
	}
	return insertedProductImage.ID, nil
}

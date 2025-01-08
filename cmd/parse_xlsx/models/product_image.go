package models

import (
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"context"
	"time"
)

type ProductImage struct {
	ID        int64
	Product   *Product
	Path      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (pi *ProductImage) InsertToDB(db database.Service) (int64, error) {
	if pi == nil {
		panic("nil ProductImage")
	}
	if pi.Product == nil {
		panic("nil ProductImage.Product")
	}
	ctx := context.Background()
	now := time.Now().UTC()
	insertedProductImage, err := db.GetQueries().CreateProductImage(ctx, queries.CreateProductImageParams{
		ProductID: pi.Product.ID,
		Path:      pi.Path,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: constants.DT_BEGINNING,
	})
	if err != nil {
		return 0, err
	}
	return insertedProductImage.ID, nil
}

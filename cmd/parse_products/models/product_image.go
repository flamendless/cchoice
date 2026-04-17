package models

import (
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"context"
	"database/sql"
)

type ProductImage struct {
	Product   *Product
	Path      string
	Thumbnail string
	ID        int64
}

func (pi *ProductImage) InsertToDB(ctx context.Context, db database.IService) (int64, error) {
	if pi == nil {
		panic("nil ProductImage")
	}
	if pi.Product == nil {
		panic("nil ProductImage.Product")
	}
	insertedProductImage, err := db.GetQueries().CreateProductImage(ctx, queries.CreateProductImageParams{
		ProductID: pi.Product.ID,
		Path:      pi.Path,
		Thumbnail: sql.NullString{Valid: pi.Thumbnail != "", String: pi.Thumbnail},
	})
	if err != nil {
		return 0, err
	}
	return insertedProductImage.ID, nil
}

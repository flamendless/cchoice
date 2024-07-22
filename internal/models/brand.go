package models

import (
	"cchoice/internal/ctx"
	"context"
)

type Brand struct {
	ID   int64
	Name string
}

func NewBrand(brandName string) Brand {
	return Brand{
		Name: brandName,
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
	newBrand, err := ctxDB.Queries.CreateBrand(ctx, brand.Name)
	if err != nil {
		return 0, err
	}
	brand.ID = newBrand.ID
	return newBrand.ID, nil
}

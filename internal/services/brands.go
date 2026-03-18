package services

import (
	"context"

	"cchoice/internal/database"
	"cchoice/internal/encode"
)

type BrandService struct {
	encoder encode.IEncode
	dbRO    database.Service
	dbRW    database.Service
}

func NewBrandService(
	encoder encode.IEncode,
	dbRO database.Service,
	dbRW database.Service,
) *BrandService {
	return &BrandService{
		encoder: encoder,
		dbRO:    dbRO,
		dbRW:    dbRW,
	}
}

func (s *BrandService) GetNameByID(ctx context.Context, brandID string) (string, error) {
	brand, err := s.dbRO.GetQueries().GetBrandsByID(ctx, s.encoder.Decode(brandID))
	if err != nil {
		return "", err
	}
	return brand.Name, nil
}

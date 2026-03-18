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

func NewBrandService(encoder encode.IEncode, dbRO database.Service, dbRW database.Service) *BrandService {
	return &BrandService{
		encoder: encoder,
		dbRO:    dbRO,
		dbRW:    dbRW,
	}
}

func (s *BrandService) GetNameByID(ctx context.Context, brandID int64) (string, error) {
	brand, err := s.dbRO.GetQueries().GetBrandsByID(ctx, brandID)
	if err != nil {
		return "", err
	}
	return brand.Name, nil
}

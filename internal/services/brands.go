package services

import (
	"context"

	"cchoice/internal/database"
	"cchoice/internal/encode"
)

type BrandsService struct {
	encoder encode.IEncode
	dbRO    database.Service
	dbRW    database.Service
}

func NewBrandsService(encoder encode.IEncode, dbRO database.Service, dbRW database.Service) *BrandsService {
	return &BrandsService{
		encoder: encoder,
		dbRO:    dbRO,
		dbRW:    dbRW,
	}
}

func (s *BrandsService) GetNameByID(ctx context.Context, brandID int64) (string, error) {
	brand, err := s.dbRO.GetQueries().GetBrandsByID(ctx, brandID)
	if err != nil {
		return "", err
	}
	return brand.Name, nil
}

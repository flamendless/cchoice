package services

import (
	"context"

	"cchoice/internal/database"
	"cchoice/internal/encode"
	"cchoice/internal/logs"
)

type BrandService struct {
	encoder encode.IEncode
	dbRO    database.IService
	dbRW    database.IService
}

func NewBrandService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
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

func (s *BrandService) Log() {
	logs.Log().Info("[BrandService] Loaded")
}

var _ IService = (*BrandService)(nil)

package services

import (
	"context"

	"cchoice/internal/database"
)

type ProductsService struct {
	dbRO database.Service
}

func NewProductsService(dbRO database.Service) *ProductsService {
	return &ProductsService{dbRO: dbRO}
}

func (s *ProductsService) ValidateSerial(ctx context.Context, serial string) (bool, error) {
	_, err := s.dbRO.GetQueries().ValidateUniqueSerial(ctx, serial)
	if err != nil {
		return true, nil
	}
	return false, nil
}

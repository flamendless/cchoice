package services

import (
	"context"

	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
)

type ProductsService struct {
	dbRO    database.Service
	dbRW    database.Service
	encoder encode.IEncode
}

func NewProductsService(
	encoder encode.IEncode,
	dbRO database.Service,
	dbRW database.Service,
) *ProductsService {
	return &ProductsService{
		dbRO:    dbRO,
		dbRW:    dbRW,
		encoder: encoder,
	}
}

func (s *ProductsService) ValidateSerial(ctx context.Context, serial string) (bool, error) {
	_, err := s.dbRO.GetQueries().ValidateUniqueSerial(ctx, serial)
	if err != nil {
		return true, nil
	}
	return false, nil
}

func (s *ProductsService) UpdateProductStatus(ctx context.Context, productID string, status enums.ProductStatus) error {
	decodedProductID := s.encoder.Decode(productID)
	if decodedProductID == encode.INVALID {
		return errs.ErrDecode
	}

	err := s.dbRW.GetQueries().UpdateProductsStatus(ctx, queries.UpdateProductsStatusParams{
		Status: status.String(),
		ID:     decodedProductID,
	})
	return err
}

package services

import (
	"context"
	"database/sql"
	"time"

	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
)

type ProductSpecsInput struct {
	Colours, Sizes, Segmentation, PartNumber string
	Power, Capacity, ScopeOfSupply           string
}

type CreateProductInput struct {
	Serial, Name, Description string
	BrandID                   int64
	Category, Subcategory     string
	Specs                     ProductSpecsInput
	ImagePath                 string
	UnitPriceWithoutVat       int64
	UnitPriceWithVat          int64
}

type ProductService struct {
	dbRO    database.Service
	dbRW    database.Service
	encoder encode.IEncode
}

func NewProductService(
	encoder encode.IEncode,
	dbRO database.Service,
	dbRW database.Service,
) *ProductService {
	return &ProductService{
		dbRO:    dbRO,
		dbRW:    dbRW,
		encoder: encoder,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, input CreateProductInput) (*queries.TblProduct, error) {
	_, err := s.dbRO.GetQueries().GetBrandsByID(ctx, input.BrandID)
	if err != nil {
		return nil, err
	}

	categoryRow, err := s.dbRO.GetQueries().GetProductCategoryByCategoryAndSubcategory(ctx, queries.GetProductCategoryByCategoryAndSubcategoryParams{
		Category:    sql.NullString{String: input.Category, Valid: true},
		Subcategory: sql.NullString{String: input.Subcategory, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	specs, err := s.dbRW.GetQueries().CreateProductSpecs(ctx, queries.CreateProductSpecsParams{
		Colours:       sql.NullString{String: input.Specs.Colours, Valid: true},
		Sizes:         sql.NullString{String: input.Specs.Sizes, Valid: true},
		Segmentation:  sql.NullString{String: input.Specs.Segmentation, Valid: true},
		PartNumber:    sql.NullString{String: input.Specs.PartNumber, Valid: true},
		Power:         sql.NullString{String: input.Specs.Power, Valid: true},
		Capacity:      sql.NullString{String: input.Specs.Capacity, Valid: true},
		ScopeOfSupply: sql.NullString{String: input.Specs.ScopeOfSupply, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	product, err := s.dbRW.GetQueries().CreateProducts(ctx, queries.CreateProductsParams{
		Serial:                      input.Serial,
		Name:                        input.Name,
		Description:                 sql.NullString{String: input.Description, Valid: true},
		BrandID:                     input.BrandID,
		Status:                      enums.PRODUCT_STATUS_DRAFT.String(),
		ProductSpecsID:              sql.NullInt64{Int64: specs.ID, Valid: true},
		UnitPriceWithoutVat:         input.UnitPriceWithoutVat,
		UnitPriceWithVat:            input.UnitPriceWithVat,
		UnitPriceWithoutVatCurrency: constants.PHP,
		UnitPriceWithVatCurrency:    constants.PHP,
		CreatedAt:                   now,
		UpdatedAt:                   now,
		DeletedAt:                   constants.DtBeginning,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.dbRW.GetQueries().CreateProductsCategories(ctx, queries.CreateProductsCategoriesParams{
		ProductID:  product.ID,
		CategoryID: categoryRow.ID,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.dbRW.GetQueries().CreateProductImage(ctx, queries.CreateProductImageParams{
		ProductID: product.ID,
		Path:      input.ImagePath,
		Thumbnail: sql.NullString{String: input.ImagePath, Valid: true},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: constants.DtBeginning,
	})
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) ValidateSerial(ctx context.Context, serial string) (bool, error) {
	_, err := s.dbRO.GetQueries().ValidateUniqueSerial(ctx, serial)
	if err != nil {
		return true, nil
	}
	return false, nil
}

func (s *ProductService) UpdateProductStatus(ctx context.Context, productID string, status enums.ProductStatus) error {
	decodedProductID := s.encoder.Decode(productID)
	if decodedProductID == encode.INVALID {
		return errs.ErrDecode
	}

	return s.dbRW.GetQueries().UpdateProductsStatus(ctx, queries.UpdateProductsStatusParams{
		Status: status.String(),
		ID:     decodedProductID,
	})
}

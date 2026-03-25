package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"cchoice/cmd/web/models"
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
	Weight                                   string
	WeightUnit                               string
}

type CreateProductInput struct {
	Serial, Name, Description string
	BrandID                   string
	Category, Subcategory     string
	Specs                     ProductSpecsInput
	ImagePath                 string
	UnitPriceWithoutVat       int64
	UnitPriceWithVat          int64
}

type ProductService struct {
	dbRO      database.Service
	dbRW      database.Service
	encoder   encode.IEncode
	getCDNURL models.CDNURLFunc
}

func NewProductService(
	encoder encode.IEncode,
	dbRO database.Service,
	dbRW database.Service,
	cdnURLFunc models.CDNURLFunc,
) *ProductService {
	return &ProductService{
		dbRO:      dbRO,
		dbRW:      dbRW,
		encoder:   encoder,
		getCDNURL: cdnURLFunc,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, input CreateProductInput) (*queries.TblProduct, error) {
	brandID := s.encoder.Decode(input.BrandID)
	_, err := s.dbRO.GetQueries().GetBrandsByID(ctx, brandID)
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

	weightVal, weightErr := strconv.ParseFloat(input.Specs.Weight, 64)
	specs, err := s.dbRW.GetQueries().CreateProductSpecs(ctx, queries.CreateProductSpecsParams{
		Colours:       sql.NullString{String: input.Specs.Colours, Valid: true},
		Sizes:         sql.NullString{String: input.Specs.Sizes, Valid: true},
		Segmentation:  sql.NullString{String: input.Specs.Segmentation, Valid: true},
		PartNumber:    sql.NullString{String: input.Specs.PartNumber, Valid: true},
		Power:         sql.NullString{String: input.Specs.Power, Valid: true},
		Capacity:      sql.NullString{String: input.Specs.Capacity, Valid: true},
		ScopeOfSupply: sql.NullString{String: input.Specs.ScopeOfSupply, Valid: true},
		Weight:        sql.NullFloat64{Float64: weightVal, Valid: weightErr != nil},
		WeightUnit:    sql.NullString{String: input.Specs.WeightUnit, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	product, err := s.dbRW.GetQueries().CreateProducts(ctx, queries.CreateProductsParams{
		Serial:                      input.Serial,
		Name:                        input.Name,
		Description:                 sql.NullString{String: input.Description, Valid: true},
		BrandID:                     brandID,
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

	if _, err = s.dbRW.GetQueries().CreateProductsCategories(ctx, queries.CreateProductsCategoriesParams{
		ProductID:  product.ID,
		CategoryID: categoryRow.ID,
	}); err != nil {
		return nil, err
	}

	if input.ImagePath != "" {
		//TODO: Implement generation of thumbnail
		if _, err = s.dbRW.GetQueries().CreateProductImage(ctx, queries.CreateProductImageParams{
			ProductID: product.ID,
			Path:      input.ImagePath,
			Thumbnail: sql.NullString{String: input.ImagePath, Valid: true},
			CreatedAt: now,
			UpdatedAt: now,
			DeletedAt: constants.DtBeginning,
		}); err != nil {
			return nil, err
		}
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

func (s *ProductService) GetProductsForListingAdmin(
	ctx context.Context,
	search, status string,
) ([]models.AdminProductListItem, error) {
	products, err := s.dbRO.GetQueries().AdminGetProductsForListing(ctx, queries.AdminGetProductsForListingParams{
		Search: sql.NullString{String: search, Valid: search != ""},
		Status: sql.NullString{String: status, Valid: status != ""},
	})
	if err != nil {
		return nil, err
	}

	productList := make([]models.AdminProductListItem, 0, len(products))
	for _, p := range products {
		var weightStr string
		if p.Weight > 0 {
			weightStr = fmt.Sprintf("%.2f %s", p.Weight, p.WeightUnit)
		}

		productList = append(productList, models.AdminProductListItem{
			ID:            s.encoder.Encode(p.ID),
			Name:          p.Name,
			Serial:        p.Serial,
			Description:   p.Description.String,
			Brand:         p.BrandName,
			Status:        enums.ParseProductStatusToEnum(p.Status),
			ThumbnailPath: p.ThumbnailPath,
			CDNURL:        s.getCDNURL(p.ThumbnailPath),
			CDNURL1280:    s.getCDNURL(constants.ToPath1280(p.ThumbnailPath)),
			CreatedAt:     p.CreatedAt.Format(constants.DateTimeLayoutISO),
			UpdatedAt:     p.UpdatedAt.Format(constants.DateTimeLayoutISO),
			Colours:       p.Colours,
			Sizes:         p.Sizes,
			Segmentation:  p.Segmentation,
			PartNumber:    p.PartNumber,
			Power:         p.Power,
			Capacity:      p.Capacity,
			ScopeOfSupply: p.ScopeOfSupply,
			Weight:        weightStr,
			WeightUnit:    p.WeightUnit,
		})
	}

	return productList, nil
}

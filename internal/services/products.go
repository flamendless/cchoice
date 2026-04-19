package services

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
)

type ProductService struct {
	dbRO      database.IService
	dbRW      database.IService
	encoder   encode.IEncode
	getCDNURL models.CDNURLFunc
}

func NewProductService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
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
	brand, err := s.dbRO.GetQueries().GetBrandsByID(ctx, brandID)
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
		Weight:        sql.NullFloat64{Float64: weightVal, Valid: weightErr == nil},
		WeightUnit:    sql.NullString{String: input.Specs.WeightUnit, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	product, err := s.dbRW.GetQueries().CreateProducts(ctx, queries.CreateProductsParams{
		Serial:                      input.Serial,
		Name:                        input.Name,
		Description:                 sql.NullString{String: input.Description, Valid: true},
		BrandID:                     brandID,
		Status:                      enums.PRODUCT_STATUS_DRAFT.String(),
		ProductSpecsID:              sql.NullInt64{Int64: specs.ID, Valid: true},
		UnitPriceWithoutVat:         input.UnitPriceWithoutVat * 100,
		UnitPriceWithVat:            input.UnitPriceWithVat * 100,
		UnitPriceWithoutVatCurrency: constants.PHP,
		UnitPriceWithVatCurrency:    constants.PHP,
		Slug: sql.NullString{
			Valid: true,
			String: utils.ProductSlug(
				brand.Name,
				categoryRow.Category.String,
				categoryRow.Subcategory.String,
				input.Serial,
				input.Specs.Power,
			),
		},
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
		//TODO: images are not yet available by this time. This should be in the thumbnail job/service
		cdnURL := s.getCDNURL(input.ImagePath)
		cdnURLThumbnail := s.getCDNURL(constants.ToPath1280(input.ImagePath))
		if _, err = s.dbRW.GetQueries().CreateProductImage(ctx, queries.CreateProductImageParams{
			ProductID:       product.ID,
			Path:            input.ImagePath,
			Thumbnail:       sql.NullString{String: input.ImagePath, Valid: true},
			CdnUrl:          sql.NullString{String: cdnURL, Valid: cdnURL != ""},
			CdnUrlThumbnail: sql.NullString{String: cdnURLThumbnail, Valid: cdnURLThumbnail != ""},
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

func (s *ProductService) DeleteProduct(ctx context.Context, productID string) error {
	decodedProductID := s.encoder.Decode(productID)
	if decodedProductID == encode.INVALID {
		return errs.ErrDecode
	}

	return s.dbRW.GetQueries().SoftDeleteProduct(ctx, decodedProductID)
}

func (s *ProductService) GetProductsForListingAdmin(
	ctx context.Context,
	search string,
	status enums.ProductStatus,
) ([]models.AdminProductListItem, error) {
	products, err := s.dbRO.GetQueries().AdminGetProductsForListing(ctx, queries.AdminGetProductsForListingParams{
		Search: sql.NullString{String: search, Valid: search != ""},
		Status: sql.NullString{String: status.String(), Valid: status != enums.PRODUCT_STATUS_UNDEFINED},
	})
	if err != nil {
		return nil, err
	}

	productList := make([]models.AdminProductListItem, 0, len(products))
	for _, p := range products {
		price := utils.NewMoney(p.UnitPriceWithVat, p.UnitPriceWithVatCurrency)
		productList = append(productList, models.AdminProductListItem{
			ID:            s.encoder.Encode(p.ID),
			Name:          p.Name,
			Serial:        p.Serial,
			Slug:          p.Slug.String,
			Description:   p.Description.String,
			Brand:         p.BrandName,
			Price:         price.Display(),
			Category:      p.Category,
			Subcategory:   p.Subcategory,
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
			Weight:        utils.ToWeightDisplay(p.Weight, p.WeightUnit),
			WeightUnit:    p.WeightUnit,
		})
	}

	return productList, nil
}

func (s *ProductService) GetProductByIDForEdit(ctx context.Context, productID string) (*ProductForEdit, error) {
	decodedProductID := s.encoder.Decode(productID)
	if decodedProductID == encode.INVALID {
		return nil, errs.ErrDecode
	}

	product, err := s.dbRO.GetQueries().GetProductsByID(ctx, decodedProductID)
	if err != nil {
		return nil, err
	}

	return &ProductForEdit{
		ID:                          product.ID,
		Serial:                      product.Serial,
		Name:                        product.Name,
		Description:                 product.Description.String,
		BrandID:                     product.BrandID,
		BrandName:                   product.BrandName,
		Status:                      product.Status,
		Category:                    product.ProductCategory,
		Subcategory:                 product.ProductSubcategory,
		ProductSpecsID:              product.ProductSpecsID.Int64,
		UnitPriceWithoutVat:         product.UnitPriceWithoutVat,
		UnitPriceWithoutVatCurrency: product.UnitPriceWithoutVatCurrency,
		UnitPriceWithVat:            product.UnitPriceWithVat,
		UnitPriceWithVatCurrency:    product.UnitPriceWithVatCurrency,
		Specs: ProductSpecsInput{
			Colours:       product.Colours.String,
			Sizes:         product.Sizes.String,
			Segmentation:  product.Segmentation.String,
			PartNumber:    product.PartNumber.String,
			Power:         product.Power.String,
			Capacity:      product.Capacity.String,
			ScopeOfSupply: product.ScopeOfSupply.String,
			Weight:        strconv.FormatFloat(product.Weight.Float64, 'f', -1, 64),
			WeightUnit:    product.WeightUnit.String,
		},
	}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, input UpdateProductInput) error {
	productID := s.encoder.Decode(input.ProductID)
	if productID == encode.INVALID {
		return errs.ErrDecode
	}

	brandID := s.encoder.Decode(input.BrandID)
	brand, err := s.dbRO.GetQueries().GetBrandsByID(ctx, brandID)
	if err != nil {
		return err
	}

	categoryRow, err := s.dbRO.GetQueries().GetProductCategoryByCategoryAndSubcategory(ctx, queries.GetProductCategoryByCategoryAndSubcategoryParams{
		Category:    sql.NullString{String: input.Category, Valid: true},
		Subcategory: sql.NullString{String: input.Subcategory, Valid: true},
	})
	if err != nil {
		return err
	}

	existingProduct, err := s.dbRO.GetQueries().GetProductsByID(ctx, productID)
	if err != nil {
		return err
	}

	weightVal, weightErr := strconv.ParseFloat(input.Specs.Weight, 64)
	if err := s.dbRW.GetQueries().UpdateProductSpecs(ctx, queries.UpdateProductSpecsParams{
		ID:            existingProduct.ProductSpecsID.Int64,
		Colours:       sql.NullString{String: input.Specs.Colours, Valid: true},
		Sizes:         sql.NullString{String: input.Specs.Sizes, Valid: true},
		Segmentation:  sql.NullString{String: input.Specs.Segmentation, Valid: true},
		PartNumber:    sql.NullString{String: input.Specs.PartNumber, Valid: true},
		Power:         sql.NullString{String: input.Specs.Power, Valid: true},
		Capacity:      sql.NullString{String: input.Specs.Capacity, Valid: true},
		ScopeOfSupply: sql.NullString{String: input.Specs.ScopeOfSupply, Valid: true},
		Weight:        sql.NullFloat64{Float64: weightVal, Valid: weightErr == nil},
		WeightUnit:    sql.NullString{String: input.Specs.WeightUnit, Valid: true},
	}); err != nil {
		return err
	}

	if _, err := s.dbRW.GetQueries().UpdateProducts(ctx, queries.UpdateProductsParams{
		ID:                          productID,
		Name:                        input.Name,
		Description:                 sql.NullString{String: input.Description, Valid: true},
		BrandID:                     brandID,
		Status:                      input.Status,
		ProductSpecsID:              existingProduct.ProductSpecsID,
		UnitPriceWithoutVat:         input.UnitPriceWithoutVat * 100,
		UnitPriceWithVat:            input.UnitPriceWithVat * 100,
		UnitPriceWithoutVatCurrency: existingProduct.UnitPriceWithoutVatCurrency,
		UnitPriceWithVatCurrency:    existingProduct.UnitPriceWithVatCurrency,
		Slug: sql.NullString{
			Valid: true,
			String: utils.ProductSlug(
				brand.Name,
				categoryRow.Category.String,
				categoryRow.Subcategory.String,
				existingProduct.Serial,
				input.Specs.Power,
			),
		},
	}); err != nil {
		return err
	}

	categoryChanged := existingProduct.ProductCategory != input.Category || existingProduct.ProductSubcategory != input.Subcategory
	if categoryChanged {
		if err := s.dbRW.GetQueries().DeleteProductsCategories(ctx, productID); err != nil {
			return err
		}

		if _, err := s.dbRW.GetQueries().CreateProductsCategories(ctx, queries.CreateProductsCategoriesParams{
			ProductID:  productID,
			CategoryID: categoryRow.ID,
		}); err != nil {
			return err
		}
	}

	if input.ImagePath != "" {
		cdnURL := s.getCDNURL(input.ImagePath)
		cdnURLThumbnail := s.getCDNURL(constants.ToPath1280(input.ImagePath))
		if _, err = s.dbRW.GetQueries().CreateProductImage(ctx, queries.CreateProductImageParams{
			ProductID:       productID,
			Path:            input.ImagePath,
			Thumbnail:       sql.NullString{String: input.ImagePath, Valid: true},
			CdnUrl:          sql.NullString{String: cdnURL, Valid: cdnURL != ""},
			CdnUrlThumbnail: sql.NullString{String: cdnURLThumbnail, Valid: cdnURLThumbnail != ""},
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *ProductService) GetProductPage(ctx context.Context, slug string) (*models.ProductPageData, error) {
	row, err := s.dbRO.GetQueries().GetProductPage(ctx, sql.NullString{Valid: slug != "", String: slug})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	origPrice := utils.NewMoney(row.UnitPriceWithVat, row.UnitPriceWithVatCurrency)
	var salePrice int64
	var saleCurrency string
	if row.IsOnSale == 1 {
		if sp, ok := row.SalePriceWithVat.(int64); ok {
			salePrice = sp
		} else {
			salePrice = row.UnitPriceWithVat
		}
		if sc, ok := row.SalePriceWithVatCurrency.(string); ok {
			saleCurrency = sc
		} else {
			saleCurrency = row.UnitPriceWithVatCurrency
		}
	} else {
		salePrice = row.UnitPriceWithVat
		saleCurrency = row.UnitPriceWithVatCurrency
	}

	discountedPrice := utils.NewMoney(salePrice, saleCurrency)
	_, _, discountPercentage := utils.GetOrigAndDiscounted(
		row.IsOnSale,
		row.UnitPriceWithVat,
		row.UnitPriceWithVatCurrency,
		sql.NullInt64{Int64: salePrice, Valid: row.IsOnSale == 1},
		sql.NullString{String: saleCurrency, Valid: row.IsOnSale == 1},
	)

	cdnURL := row.CdnUrl
	if cdnURL == "" {
		cdnURL = s.getCDNURL(row.ThumbnailPath)
	}
	cdnURL1280 := row.CdnUrlThumbnail
	if cdnURL1280 == "" {
		cdnURL1280 = s.getCDNURL(constants.ToPath1280(row.ThumbnailPath))
	}

	colours := strings.Split(row.Colours, ",")
	sizes := strings.Split(row.Sizes, ",")

	specs := make(map[string]string)
	if row.Segmentation != "" {
		specs["Segmentation"] = row.Segmentation
	}
	if row.PartNumber != "" {
		specs["Part Number"] = row.PartNumber
	}
	if row.Power != "" {
		specs["Power"] = row.Power
	}
	if row.Capacity != "" {
		specs["Capacity"] = row.Capacity
	}
	if row.ScopeOfSupply != "" {
		specs["Scope of Supply"] = row.ScopeOfSupply
	}
	if row.Weight > 0 {
		specs["Weight"] = utils.ToWeightDisplay(row.Weight, row.WeightUnit)
	}

	return &models.ProductPageData{
		ProductID:                  s.encoder.Encode(row.ID),
		Serial:                     row.Serial,
		Name:                       row.Name,
		Description:                row.Description.String,
		BrandID:                    s.encoder.Encode(row.BrandID),
		BrandName:                  row.BrandName,
		BrandThumbnail:             row.BrandThumbnailUrl.String,
		ProductCategory:            row.ProductCategory,
		ProductSubcategory:         row.ProductSubcategory,
		ImagePath:                  row.ImagePath,
		ThumbnailPath:              row.ThumbnailPath,
		CDNURL:                     cdnURL,
		CDNURL1280:                 cdnURL1280,
		UnitPriceWithoutVat:        row.UnitPriceWithoutVat,
		UnitPriceWithVat:           row.UnitPriceWithVat,
		UnitPriceWithoutVatDisplay: origPrice.Display(),
		PriceDisplay:               discountedPrice.Display(),
		OrigPriceDisplay:           origPrice.Display(),
		DiscountPercentage:         discountPercentage,
		IsOnSale:                   row.IsOnSale == 1,
		Colours:                    colours,
		Sizes:                      sizes,
		Specs:                      specs,
	}, nil
}

func (s *ProductService) ID() string {
	return "Product"
}

func (s *ProductService) Log() {
	logs.Log().Info("[ProductService] Loaded")
}

var _ IService = (*ProductService)(nil)

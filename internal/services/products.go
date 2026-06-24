package services

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

type ProductService struct {
	dbRO             database.IService
	dbRW             database.IService
	encoder          encode.IEncode
	getCDNURL        models.CDNURLFunc
	productInventory *ProductInventoryService
	staffLog         *StaffLogsService
}

func NewProductService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	cdnURLFunc models.CDNURLFunc,
	productInventory *ProductInventoryService,
	staffLog *StaffLogsService,
) *ProductService {
	if productInventory == nil {
		panic("ProductInventoryService is required")
	}
	if staffLog == nil {
		panic("StaffLogsService is required")
	}
	return &ProductService{
		dbRO:             dbRO,
		dbRW:             dbRW,
		encoder:          encoder,
		getCDNURL:        cdnURLFunc,
		productInventory: productInventory,
		staffLog:         staffLog,
	}
}

func (s *ProductService) Create(
	ctx context.Context,
	staffID string,
	input CreateProductInput,
) (*queries.TblProduct, error) {
	staffDBID := s.encoder.Decode(staffID)
	if staffDBID == encode.INVALID {
		return nil, errs.ErrDecode
	}

	brandID := s.encoder.Decode(input.BrandID)
	if brandID == encode.INVALID {
		return nil, errs.ErrDecode
	}

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

	if _, err := s.productInventory.Create(
		ctx,
		staffID,
		s.encoder.Encode(product.ID),
		input.Stocks,
		input.StocksIn,
	); err != nil {
		return nil, err
	}

	if input.SalePriceWithVat > 0 {
		if err := s.syncProductSale(
			ctx,
			product.ID,
			input.UnitPriceWithVat,
			input.SalePriceWithoutVat,
			input.SalePriceWithVat,
			input.SaleStartDate,
			input.SaleEndDate,
		); err != nil {
			return nil, err
		}
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

func (s *ProductService) UpdateStatus(ctx context.Context, productID string, status enums.ProductStatus) error {
	decodedProductID := s.encoder.Decode(productID)
	if decodedProductID == encode.INVALID {
		return errs.ErrDecode
	}

	return s.dbRW.GetQueries().UpdateProductsStatus(ctx, queries.UpdateProductsStatusParams{
		Status: status.String(),
		ID:     decodedProductID,
	})
}

func (s *ProductService) Delete(ctx context.Context, productID string) error {
	decodedProductID := s.encoder.Decode(productID)
	if decodedProductID == encode.INVALID {
		return errs.ErrDecode
	}

	return s.dbRW.GetQueries().SoftDeleteProduct(ctx, decodedProductID)
}

func (s *ProductService) GetForListingAdmin(
	ctx context.Context,
	searchSerial string,
	searchBrand string,
	status enums.ProductStatus,
) ([]models.AdminProductListItem, error) {
	products, err := s.dbRO.GetQueries().AdminGetProductsForListing(ctx, queries.AdminGetProductsForListingParams{
		SearchSerial: sql.NullString{String: searchSerial, Valid: searchSerial != ""},
		SearchBrand:  sql.NullString{String: searchBrand, Valid: searchBrand != ""},
		Status:       sql.NullString{String: status.String(), Valid: status != enums.PRODUCT_STATUS_UNDEFINED},
	})
	if err != nil {
		return nil, err
	}

	productList := make([]models.AdminProductListItem, 0, len(products))
	for _, p := range products {
		price := utils.NewMoney(p.UnitPriceWithVat, p.UnitPriceWithVatCurrency)
		productID := s.encoder.Encode(p.ID)
		inventory, err := s.productInventory.GetByProductID(ctx, productID)
		if err != nil || inventory == nil {
			err = cmp.Or(err, errs.ErrDBNil)
			logs.Log().Warn(s.ID(), zap.String("product id", productID), zap.Error(err))
			continue
		}

		cdnURL := p.CdnUrl.String
		if cdnURL == "" {
			cdnURL = s.getCDNURL(p.ThumbnailPath)
		}
		cdnURL1280 := p.CdnUrlThumbnail.String
		if cdnURL1280 == "" {
			cdnURL1280 = s.getCDNURL(constants.ToPath1280(p.ThumbnailPath))
		}

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
			CDNURL:        cdnURL,
			CDNURL1280:    cdnURL1280,
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
			Stocks:        strconv.FormatInt(inventory.Stocks, 10),
		})
	}

	return productList, nil
}

func (s *ProductService) GetForExportAdmin(
	ctx context.Context,
	brand string,
	status enums.ProductStatus,
	sortColumn enums.ProductExportSortColumn,
	sortDirection enums.ProductExportSortDirection,
) ([]ProductExportRow, error) {
	products, err := s.dbRO.GetQueries().AdminGetProductsForExport(ctx, queries.AdminGetProductsForExportParams{
		SearchBrand: sql.NullString{String: brand, Valid: brand != ""},
		Status:      sql.NullString{String: status.String(), Valid: status != enums.PRODUCT_STATUS_UNDEFINED},
	})
	if err != nil {
		return nil, err
	}

	rows := make([]ProductExportRow, 0, len(products))
	for _, p := range products {
		imageURL := p.CdnUrl.String
		if imageURL == "" && p.ImagePath != "" {
			imageURL = s.getCDNURL(p.ImagePath)
		}
		if imageURL == "" {
			imageURL = p.ImagePath
		}

		thumbnailURL := p.CdnUrlThumbnail.String
		if thumbnailURL == "" && p.ThumbnailPath != "" {
			thumbnailURL = s.getCDNURL(constants.ToPath1280(p.ThumbnailPath))
		}
		if thumbnailURL == "" {
			thumbnailURL = p.ThumbnailPath
		}

		unitPrice := utils.NewMoney(p.UnitPriceWithVat, p.UnitPriceWithVatCurrency).Display()

		salePrice := ""
		if p.SalePriceWithVat.Valid {
			currency := p.SalePriceWithVatCurrency.String
			if currency == "" {
				currency = p.UnitPriceWithVatCurrency
			}
			salePrice = utils.NewMoney(p.SalePriceWithVat.Int64, currency).Display()
		}

		saleStartDate := ""
		if p.SaleStartsAt.Valid {
			saleStartDate = p.SaleStartsAt.Time.Format(constants.DateTimeLayoutISO)
		}
		saleEndDate := ""
		if p.SaleEndsAt.Valid {
			saleEndDate = p.SaleEndsAt.Time.Format(constants.DateTimeLayoutISO)
		}

		stocksIn := ""
		stocksQty := ""
		if p.StocksIn.Valid {
			stocksIn = p.StocksIn.String
		}
		if p.Stocks.Valid {
			stocksQty = strconv.FormatInt(p.Stocks.Int64, 10)
		}

		weightStr := ""
		if p.Weight != 0 {
			weightStr = strconv.FormatFloat(p.Weight, 'f', -1, 64)
		}

		rows = append(rows, ProductExportRow{
			Brand:              p.BrandName,
			Serial:             p.Serial,
			Slug:               p.Slug.String,
			Status:             p.Status,
			Category:           p.Category,
			Subcategory:        p.Subcategory,
			Name:               p.Name,
			UnitPriceWithVat:   unitPrice,
			SalePriceWithVat:   salePrice,
			SaleStartDate:      saleStartDate,
			SaleEndDate:        saleEndDate,
			Description:        p.Description.String,
			Colours:            p.Colours,
			Sizes:              p.Sizes,
			Segmentation:       p.Segmentation,
			PartNumber:         p.PartNumber,
			Power:              p.Power,
			Capacity:           p.Capacity,
			Weight:             weightStr,
			WeightUnit:         p.WeightUnit,
			ScopeOfSupply:      p.ScopeOfSupply,
			StocksIn:           stocksIn,
			StocksQty:          stocksQty,
			ImageURL:     imageURL,
			ThumbnailURL: thumbnailURL,
			CreatedAt:    p.CreatedAt.Format(constants.DateTimeLayoutISO),
			UpdatedAt:    p.UpdatedAt.Format(constants.DateTimeLayoutISO),
		})
	}

	sortProductExportRows(rows, sortColumn, sortDirection)

	return rows, nil
}

func (s *ProductService) CountForExportAdmin(
	ctx context.Context,
	brand string,
	status enums.ProductStatus,
) (int64, error) {
	count, err := s.dbRO.GetQueries().AdminCountProductsForExport(ctx, queries.AdminCountProductsForExportParams{
		SearchBrand: sql.NullString{String: brand, Valid: brand != ""},
		Status:      sql.NullString{String: status.String(), Valid: status != enums.PRODUCT_STATUS_UNDEFINED},
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *ProductService) GetByIDForEdit(ctx context.Context, productID string) (*ProductForEdit, error) {
	decodedProductID := s.encoder.Decode(productID)
	if decodedProductID == encode.INVALID {
		return nil, errs.ErrDecode
	}

	product, err := s.dbRO.GetQueries().GetProductsByID(ctx, decodedProductID)
	if err != nil {
		return nil, err
	}

	inventory, err := s.productInventory.GetByProductID(ctx, productID)
	if err != nil || inventory == nil {
		return nil, cmp.Or(err, errs.ErrDBNil)
	}

	var salePriceWithVat int64
	var saleStartDate, saleEndDate string
	if sale, saleErr := s.dbRO.GetQueries().GetActiveSaleByProductID(ctx, decodedProductID); saleErr == nil {
		salePriceWithVat = sale.SalePriceWithVat
		saleStartDate = sale.StartsAt.Format(constants.DateLayoutISO)
		saleEndDate = sale.EndsAt.Format(constants.DateLayoutISO)
	} else if saleErr != sql.ErrNoRows {
		return nil, saleErr
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
		SalePriceWithVat:            salePriceWithVat,
		SaleStartDate:               saleStartDate,
		SaleEndDate:                 saleEndDate,
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
		StocksIn: inventory.StocksIn,
		Stocks:   inventory.Stocks,
	}, nil
}

func (s *ProductService) Update(ctx context.Context, staffID string, input UpdateProductInput) error {
	staffIDDB := s.encoder.Decode(staffID)
	if staffIDDB == encode.INVALID {
		return errs.ErrDecode
	}

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

	if err := s.productInventory.SetQty(ctx, staffID, input.ProductID, input.Stocks, input.StocksIn); err != nil {
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
		if !existingProduct.ProductImageID.Valid || existingProduct.ProductImageID.Int64 == 0 {
			if _, err = s.dbRW.GetQueries().CreateProductImage(ctx, queries.CreateProductImageParams{
				ProductID:       productID,
				Path:            input.ImagePath,
				Thumbnail:       sql.NullString{String: input.ImagePath, Valid: true},
				CdnUrl:          sql.NullString{String: cdnURL, Valid: cdnURL != ""},
				CdnUrlThumbnail: sql.NullString{String: cdnURLThumbnail, Valid: cdnURLThumbnail != ""},
			}); err != nil {
				return err
			}
		} else {
			if err = s.dbRW.GetQueries().UpdateProductImage(ctx, queries.UpdateProductImageParams{
				ID:              existingProduct.ProductImageID.Int64,
				Path:            input.ImagePath,
				Thumbnail:       sql.NullString{String: input.ImagePath, Valid: true},
				CdnUrl:          sql.NullString{String: cdnURL, Valid: cdnURL != ""},
				CdnUrlThumbnail: sql.NullString{String: cdnURLThumbnail, Valid: cdnURLThumbnail != ""},
			}); err != nil {
				return err
			}
		}
	}

	if err := s.syncProductSale(
		ctx,
		productID,
		input.UnitPriceWithVat,
		input.SalePriceWithoutVat,
		input.SalePriceWithVat,
		input.SaleStartDate,
		input.SaleEndDate,
	); err != nil {
		return err
	}

	return nil
}

func productSaleDiscountValue(unitPriceWithVatPesos, salePriceWithVatPesos int64) int64 {
	discountValue := unitPriceWithVatPesos*100 - salePriceWithVatPesos*100
	if discountValue < 0 {
		return 0
	}
	return discountValue
}

func (s *ProductService) syncProductSale(
	ctx context.Context,
	productID int64,
	unitPriceWithVat int64,
	salePriceWithoutVat int64,
	salePriceWithVat int64,
	saleStartDate string,
	saleEndDate string,
) error {
	existingSale, err := s.dbRO.GetQueries().GetActiveSaleByProductID(ctx, productID)
	hasActiveSale := err == nil
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if salePriceWithVat > 0 {
		startsAt, startErr := time.Parse(constants.DateLayoutISO, saleStartDate)
		if startErr != nil {
			return fmt.Errorf("invalid sale start date: %w", startErr)
		}
		endsAt, endErr := time.Parse(constants.DateLayoutISO, saleEndDate)
		if endErr != nil {
			return fmt.Errorf("invalid sale end date: %w", endErr)
		}

		discountValue := productSaleDiscountValue(unitPriceWithVat, salePriceWithVat)
		if hasActiveSale {
			return s.dbRW.GetQueries().UpdateProductSale(ctx, queries.UpdateProductSaleParams{
				SalePriceWithoutVat:         salePriceWithoutVat * 100,
				SalePriceWithVat:            salePriceWithVat * 100,
				SalePriceWithoutVatCurrency: constants.PHP,
				SalePriceWithVatCurrency:    constants.PHP,
				DiscountType:                "fixed",
				DiscountValue:               discountValue,
				StartsAt:                    startsAt,
				EndsAt:                      endsAt,
				ID:                          existingSale.ID,
			})
		}

		_, err := s.dbRW.GetQueries().CreateProductSale(ctx, queries.CreateProductSaleParams{
			ProductID:                   productID,
			SalePriceWithoutVat:         salePriceWithoutVat * 100,
			SalePriceWithVat:            salePriceWithVat * 100,
			SalePriceWithoutVatCurrency: constants.PHP,
			SalePriceWithVatCurrency:    constants.PHP,
			DiscountType:                "fixed",
			DiscountValue:               discountValue,
			StartsAt:                    startsAt,
			EndsAt:                      endsAt,
			IsActive:                    true,
		})
		return err
	}

	if hasActiveSale {
		return s.dbRW.GetQueries().DeactivateProductSalesByProductID(ctx, productID)
	}

	return nil
}

func (s *ProductService) GetForPage(ctx context.Context, slug string) (*models.ProductPageData, error) {
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

	specs := []models.ProductSpec{
		models.ProductSpec{Label: "Segmentation", Value: row.Segmentation},
		models.ProductSpec{Label: "Part Number", Value: row.PartNumber},
		models.ProductSpec{Label: "Power", Value: row.Power},
		models.ProductSpec{Label: "Capacity", Value: row.Capacity},
		models.ProductSpec{Label: "Scope of Supply", Value: row.ScopeOfSupply},
		models.ProductSpec{Label: "Weight", Value: utils.ToWeightDisplay(row.Weight, row.WeightUnit)},
	}

	return &models.ProductPageData{
		Meta:                       s.GenerateMeta(&row),
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

func (s *ProductService) GenerateMeta(product *queries.GetProductPageRow) models.ProductsMeta {
	title := fmt.Sprintf(
		"%s %s %s (%s) - Price, Specs, Buy Online",
		product.BrandName,
		product.Name,
		product.Power,
		product.Serial,
	)
	content := fmt.Sprintf(
		"Buy %s %s %s (%s). Power Tool. Durable. Quality. Best Prices in the Philippines. Free Shipping Available. %s",
		product.BrandName,
		product.Name,
		product.Power,
		product.Serial,
		product.Description.String,
	)
	return models.ProductsMeta{
		Title:   title,
		Content: content,
	}
}

func (s *ProductService) ListForQuotations(ctx context.Context) ([]queries.ListProductsForQuotationsRow, error) {
	products, err := s.dbRO.GetQueries().ListProductsForQuotations(ctx) // TODO: paginations
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *ProductService) GetAllCategoryNames(ctx context.Context) ([]string, error) {
	categories, err := s.dbRO.GetQueries().GetAllCategoryNames(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(categories))
	for _, c := range categories {
		if c.Valid {
			names = append(names, c.String)
		}
	}
	return names, nil
}

func (s *ProductService) GetAllSubcategoryNames(ctx context.Context) ([]string, error) {
	subcategories, err := s.dbRO.GetQueries().GetAllSubcategoryNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all subcategory names: %w", err)
	}
	names := make([]string, 0, len(subcategories))
	for _, s := range subcategories {
		if s.Valid {
			names = append(names, s.String)
		}
	}
	return names, nil
}

func (s *ProductService) ID() string {
	return "Product"
}

func (s *ProductService) Log() {
	logs.Log().Info("[ProductService] Loaded")
}

var _ IService = (*ProductService)(nil)

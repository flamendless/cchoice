package services

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
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
	products, err := s.dbRO.GetQueries().AdminGetProductsForListing(ctx, s.listingAdminFilterParams(searchSerial, searchBrand, status))
	if err != nil {
		return nil, err
	}

	return s.mapAdminProductListItems(products), nil
}

func (s *ProductService) GetForListingAdminPaginated(
	ctx context.Context,
	searchSerial string,
	searchBrand string,
	status enums.ProductStatus,
	page, perPage int,
) ([]models.AdminProductListItem, int64, int, error) {
	filterParams := s.listingAdminFilterParams(searchSerial, searchBrand, status)

	totalCount, err := s.dbRO.GetQueries().AdminCountProductsForListing(ctx, queries.AdminCountProductsForListingParams(filterParams))
	if err != nil {
		return nil, 0, 0, err
	}

	page = models.ClampPage(page, totalCount, perPage)
	offset := int64((page - 1) * perPage)

	products, err := s.dbRO.GetQueries().AdminGetProductsForListingPaginated(ctx, queries.AdminGetProductsForListingPaginatedParams{
		SearchSerial: filterParams.SearchSerial,
		SearchBrand:  filterParams.SearchBrand,
		Status:       filterParams.Status,
		Limit:        int64(perPage),
		Offset:       offset,
	})
	if err != nil {
		return nil, 0, 0, err
	}

	return s.mapAdminProductListItemsPaginated(products), totalCount, page, nil
}

func (s *ProductService) listingAdminFilterParams(
	searchSerial string,
	searchBrand string,
	status enums.ProductStatus,
) queries.AdminGetProductsForListingParams {
	return queries.AdminGetProductsForListingParams{
		SearchSerial: sql.NullString{String: searchSerial, Valid: searchSerial != ""},
		SearchBrand:  sql.NullString{String: searchBrand, Valid: searchBrand != ""},
		Status:       sql.NullString{String: status.String(), Valid: status != enums.PRODUCT_STATUS_UNDEFINED},
	}
}

func (s *ProductService) mapAdminProductListItems(products []queries.AdminGetProductsForListingRow) []models.AdminProductListItem {
	productList := make([]models.AdminProductListItem, 0, len(products))
	for _, p := range products {
		productList = append(productList, s.mapAdminProductListItem(
			p.ID, p.Name, p.Serial, p.Slug, p.Description, p.BrandName, p.Status,
			p.UnitPriceWithVat, p.UnitPriceWithVatCurrency, p.CreatedAt, p.UpdatedAt,
			p.ThumbnailPath, p.CdnUrl, p.CdnUrlThumbnail, p.Category, p.Subcategory,
			p.SalePriceWithVat, p.SalePriceWithVatCurrency,
			p.Colours, p.Sizes, p.Segmentation, p.PartNumber, p.Power, p.Capacity, p.ScopeOfSupply,
			p.Weight, p.WeightUnit,
		))
	}
	return productList
}

func (s *ProductService) mapAdminProductListItemsPaginated(products []queries.AdminGetProductsForListingPaginatedRow) []models.AdminProductListItem {
	productList := make([]models.AdminProductListItem, 0, len(products))
	for _, p := range products {
		productList = append(productList, s.mapAdminProductListItem(
			p.ID, p.Name, p.Serial, p.Slug, p.Description, p.BrandName, p.Status,
			p.UnitPriceWithVat, p.UnitPriceWithVatCurrency, p.CreatedAt, p.UpdatedAt,
			p.ThumbnailPath, p.CdnUrl, p.CdnUrlThumbnail, p.Category, p.Subcategory,
			p.SalePriceWithVat, p.SalePriceWithVatCurrency,
			p.Colours, p.Sizes, p.Segmentation, p.PartNumber, p.Power, p.Capacity, p.ScopeOfSupply,
			p.Weight, p.WeightUnit,
		))
	}
	return productList
}

func (s *ProductService) mapAdminProductListItem(
	id int64,
	name, serial string,
	slug, description sql.NullString,
	brandName, status string,
	unitPriceWithVat int64,
	unitPriceWithVatCurrency string,
	createdAt, updatedAt time.Time,
	thumbnailPath string,
	cdnUrl, cdnUrlThumbnail sql.NullString,
	category, subcategory string,
	salePriceWithVat sql.NullInt64,
	salePriceWithVatCurrency sql.NullString,
	colours, sizes, segmentation, partNumber, power, capacity, scopeOfSupply string,
	weight float64,
	weightUnit string,
) models.AdminProductListItem {
	price := utils.NewMoney(unitPriceWithVat, unitPriceWithVatCurrency)

	salePrice := ""
	if salePriceWithVat.Valid {
		currency := salePriceWithVatCurrency.String
		if currency == "" {
			currency = unitPriceWithVatCurrency
		}
		salePrice = utils.NewMoney(salePriceWithVat.Int64, currency).Display()
	}

	cdnURL := cdnUrl.String
	if cdnURL == "" {
		cdnURL = s.getCDNURL(thumbnailPath)
	}
	cdnURL1280 := cdnUrlThumbnail.String
	if cdnURL1280 == "" {
		cdnURL1280 = s.getCDNURL(constants.ToPath1280(thumbnailPath))
	}

	return models.AdminProductListItem{
		ID:            s.encoder.Encode(id),
		Name:          name,
		Serial:        serial,
		Slug:          slug.String,
		Description:   description.String,
		Brand:         brandName,
		Price:         price.Display(),
		SalePrice:     salePrice,
		Category:      category,
		Subcategory:   subcategory,
		Status:        enums.ParseProductStatusToEnum(status),
		ThumbnailPath: thumbnailPath,
		CDNURL:        cdnURL,
		CDNURL1280:    cdnURL1280,
		CreatedAt:     createdAt.Format(constants.DateTimeLayoutISO),
		UpdatedAt:     updatedAt.Format(constants.DateTimeLayoutISO),
		Colours:       colours,
		Sizes:         sizes,
		Segmentation:  segmentation,
		PartNumber:    partNumber,
		Power:         power,
		Capacity:      capacity,
		ScopeOfSupply: scopeOfSupply,
		Weight:        utils.ToWeightDisplay(weight, weightUnit),
		WeightUnit:    weightUnit,
	}
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

	relatedProducts, err := s.getRelatedProducts(ctx, row.CategoryID, row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get related products: %w", err)
	}

	priceAmount, priceCurrency := utils.SchemaPrice(salePrice, saleCurrency)
	productSlug := slug
	if row.Slug.Valid && row.Slug.String != "" {
		productSlug = row.Slug.String
	}
	meta := s.GenerateMeta(&row, productSlug, cdnURL1280, priceAmount, priceCurrency)

	return &models.ProductPageData{
		Meta:                       meta,
		Slug:                       productSlug,
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
		RelatedProducts:            relatedProducts,
	}, nil
}

func (s *ProductService) getRelatedProducts(
	ctx context.Context,
	categoryID sql.NullInt64,
	productID int64,
) ([]models.RelatedProduct, error) {
	if !categoryID.Valid || categoryID.Int64 == 0 {
		return nil, nil
	}

	rows, err := s.dbRO.GetQueries().GetRelatedProductsByCategory(ctx, queries.GetRelatedProductsByCategoryParams{
		CategoryID: categoryID.Int64,
		ID:         productID,
	})
	if err != nil {
		return nil, err
	}

	related := make([]models.RelatedProduct, 0, len(rows))
	for _, row := range rows {
		if !row.Slug.Valid || row.Slug.String == "" {
			continue
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

		related = append(related, models.RelatedProduct{
			ProductID:          s.encoder.Encode(row.ID),
			Slug:               row.Slug.String,
			Name:               row.Name,
			Serial:             row.Serial,
			BrandName:          row.BrandName,
			CDNURL:             cdnURL,
			CDNURL1280:         cdnURL1280,
			OrigPriceDisplay:   origPrice.Display(),
			PriceDisplay:       discountedPrice.Display(),
			DiscountPercentage: discountPercentage,
		})
	}

	return related, nil
}

func (s *ProductService) GenerateMeta(
	product *queries.GetProductPageRow,
	slug string,
	imageURL string,
	priceAmount string,
	priceCurrency string,
) models.ProductsMeta {
	title := fmt.Sprintf(
		"%s %s (%s) - Price, Specs, Buy Online | C-Choice",
		product.BrandName,
		product.Name,
		product.Serial,
	)

	description := strings.TrimSpace(product.Description.String)
	if description == "" {
		description = fmt.Sprintf(
			"Shop %s %s (%s) from %s. Quality construction supplies with competitive pricing in the Philippines.",
			product.BrandName,
			product.Name,
			product.Serial,
			product.BrandName,
		)
	} else if len(description) > 155 {
		description = description[:152] + "..."
	}

	keywords := strings.Join([]string{
		product.BrandName,
		product.Name,
		product.Serial,
		product.ProductCategory,
		product.ProductSubcategory,
		"c-choice",
		"construction supplies",
		"philippines",
	}, ", ")

	canonicalURL := utils.SiteURL("/product/" + slug)
	ogImage := imageURL
	if ogImage == "" {
		ogImage = models.DefaultSiteSEO().OGImage
	}

	return models.ProductsMeta{
		Title:          title,
		Content:        description,
		CanonicalURL:   canonicalURL,
		OGImage:        ogImage,
		OGType:         "product",
		Robots:         "index, follow, max-image-preview:large",
		Keywords:       keywords,
		TwitterCard:    "summary_large_image",
		PriceAmount:    priceAmount,
		PriceCurrency:  priceCurrency,
		StructuredData: buildProductStructuredData(product, canonicalURL, ogImage, priceAmount, priceCurrency),
	}
}

func buildProductStructuredData(
	product *queries.GetProductPageRow,
	canonicalURL string,
	imageURL string,
	priceAmount string,
	priceCurrency string,
) string {
	type brand struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	}
	type offer struct {
		Type         string `json:"@type"`
		URL          string `json:"url"`
		PriceCurrency string `json:"priceCurrency"`
		Price        string `json:"price"`
		Availability string `json:"availability"`
		ItemCondition string `json:"itemCondition"`
	}
	type breadcrumbItem struct {
		Type     string `json:"@type"`
		Position int    `json:"position"`
		Name     string `json:"name"`
		Item     string `json:"item,omitempty"`
	}
	type breadcrumbList struct {
		Type     string           `json:"@type"`
		ItemList []breadcrumbItem `json:"itemListElement"`
	}
	type productSchema struct {
		Context     string         `json:"@context"`
		Type        string         `json:"@type"`
		Name        string         `json:"name"`
		Description string         `json:"description,omitempty"`
		Image       []string       `json:"image,omitempty"`
		SKU         string         `json:"sku"`
		Brand       brand          `json:"brand"`
		Offers      offer          `json:"offers"`
		Breadcrumb  breadcrumbList `json:"breadcrumb"`
	}

	description := strings.TrimSpace(product.Description.String)
	images := []string{}
	if imageURL != "" {
		images = append(images, imageURL)
	}

	items := []breadcrumbItem{
		{Type: "ListItem", Position: 1, Name: "Home", Item: utils.SiteURL("/")},
	}
	position := 2
	if product.ProductCategory != "" {
		items = append(items, breadcrumbItem{
			Type:     "ListItem",
			Position: position,
			Name:     product.ProductCategory,
		})
		position++
	}
	if product.ProductSubcategory != "" {
		items = append(items, breadcrumbItem{
			Type:     "ListItem",
			Position: position,
			Name:     product.ProductSubcategory,
		})
		position++
	}
	items = append(items, breadcrumbItem{
		Type:     "ListItem",
		Position: position,
		Name:     product.Name,
		Item:     canonicalURL,
	})

	schema := productSchema{
		Context:     "https://schema.org",
		Type:        "Product",
		Name:        fmt.Sprintf("%s %s", product.BrandName, product.Name),
		Description: description,
		Image:       images,
		SKU:         product.Serial,
		Brand:       brand{Type: "Brand", Name: product.BrandName},
		Offers: offer{
			Type:          "Offer",
			URL:           canonicalURL,
			PriceCurrency: priceCurrency,
			Price:         priceAmount,
			Availability:  "https://schema.org/InStock",
			ItemCondition: "https://schema.org/NewCondition",
		},
		Breadcrumb: breadcrumbList{
			Type:     "BreadcrumbList",
			ItemList: items,
		},
	}

	data, err := json.Marshal(schema)
	if err != nil {
		return ""
	}
	return string(data)
}

func (s *ProductService) ListActiveProductSlugs(ctx context.Context) ([]queries.ListActiveProductSlugsRow, error) {
	return s.dbRO.GetQueries().ListActiveProductSlugs(ctx)
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

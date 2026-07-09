package services

import (
	"context"
	"database/sql"
	"strings"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/utils"
)

const (
	SearchRelatedSourceSubcategory = "related"
	SearchRelatedSourceCategory    = "category"
	SearchRelatedSourceBrand       = "brand"
)

type SearchRelatedProductsResult struct {
	Products []models.CategorySectionProduct
	Source   string
	HasMore  bool
}

type relatedProductQueryRow struct {
	ID                       int64
	Slug                     sql.NullString
	Name                     string
	Serial                   string
	UnitPriceWithVat         int64
	UnitPriceWithVatCurrency string
	BrandName                string
	ThumbnailPath            string
	CdnUrl                   string
	CdnUrlThumbnail          string
	IsOnSale                 int64
	SalePriceWithVat         any
	SalePriceWithVatCurrency any
}

func (s *ProductService) getRelatedProducts(
	ctx context.Context,
	product queries.GetProductPageRow,
) ([]models.RelatedProduct, error) {
	if product.CategoryID.Valid && product.CategoryID.Int64 != 0 {
		rows, err := s.dbRO.GetQueries().GetRelatedProductsByCategory(ctx, queries.GetRelatedProductsByCategoryParams{
			CategoryID: product.CategoryID.Int64,
			ID:         product.ID,
		})
		if err != nil {
			return nil, err
		}
		if related := s.mapRelatedProductQueryRows(toRelatedProductQueryRowsFromCategory(rows)); len(related) > 0 {
			return related, nil
		}
	}

	if product.ProductCategory != "" {
		rows, err := s.dbRO.GetQueries().GetRelatedProductsByParentCategory(ctx, queries.GetRelatedProductsByParentCategoryParams{
			Category: sql.NullString{String: product.ProductCategory, Valid: product.ProductCategory != ""},
			ID:       product.ID,
		})
		if err != nil {
			return nil, err
		}
		if related := s.mapRelatedProductQueryRows(toRelatedProductQueryRowsFromParentCategory(rows)); len(related) > 0 {
			return related, nil
		}
	}

	rows, err := s.dbRO.GetQueries().GetRelatedProductsByBrand(ctx, queries.GetRelatedProductsByBrandParams{
		BrandID: product.BrandID,
		ID:      product.ID,
	})
	if err != nil {
		return nil, err
	}

	return s.mapRelatedProductQueryRows(toRelatedProductQueryRowsFromBrand(rows)), nil
}

func (s *ProductService) GetRelatedProductsForPage(ctx context.Context, slug string) ([]models.RelatedProduct, error) {
	row, err := s.dbRO.GetQueries().GetProductPage(ctx, sql.NullString{Valid: slug != "", String: slug})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return s.getRelatedProducts(ctx, row)
}

func (s *ProductService) GetSearchRelatedProducts(
	ctx context.Context,
	query string,
	source string,
	page int,
) (SearchRelatedProductsResult, error) {
	limit := constants.DefaultLimitSearchResultsPage
	offset := page * limit

	if page == 0 {
		for _, tier := range []string{
			SearchRelatedSourceSubcategory,
			SearchRelatedSourceCategory,
			SearchRelatedSourceBrand,
		} {
			rows, err := s.fetchSearchRelatedRows(ctx, query, tier, int64(limit), int64(offset))
			if err != nil {
				return SearchRelatedProductsResult{}, err
			}

			validRows := filterSearchRelatedRows(rows)
			if len(validRows) > 0 {
				return SearchRelatedProductsResult{
					Products: models.ToProductGridProductsFromRelatedRows(s.encoder, s.getCDNURL, validRows),
					Source:   tier,
					HasMore:  len(validRows) == limit,
				}, nil
			}
		}

		return SearchRelatedProductsResult{}, nil
	}

	if source != SearchRelatedSourceSubcategory &&
		source != SearchRelatedSourceCategory &&
		source != SearchRelatedSourceBrand {
		source = SearchRelatedSourceSubcategory
	}

	rows, err := s.fetchSearchRelatedRows(ctx, query, source, int64(limit), int64(offset))
	if err != nil {
		return SearchRelatedProductsResult{}, err
	}

	validRows := filterSearchRelatedRows(rows)
	return SearchRelatedProductsResult{
		Products: models.ToProductGridProductsFromRelatedRows(s.encoder, s.getCDNURL, validRows),
		Source:   source,
		HasMore:  len(validRows) == limit,
	}, nil
}

func (s *ProductService) fetchSearchRelatedRows(
	ctx context.Context,
	query string,
	source string,
	limit int64,
	offset int64,
) ([]queries.GetRelatedProductsForSearchRow, error) {
	params := queries.GetRelatedProductsForSearchParams{
		SearchQuery: query,
		Limit:       limit,
		Offset:      offset,
	}

	switch source {
	case SearchRelatedSourceCategory:
		rows, err := s.dbRO.GetQueries().GetRelatedProductsForSearchByParentCategory(ctx, queries.GetRelatedProductsForSearchByParentCategoryParams(params))
		if err != nil {
			return nil, err
		}
		return toSearchRelatedRowsFromParentCategory(rows), nil
	case SearchRelatedSourceBrand:
		rows, err := s.dbRO.GetQueries().GetRelatedProductsForSearchByBrand(ctx, queries.GetRelatedProductsForSearchByBrandParams(params))
		if err != nil {
			return nil, err
		}
		return toSearchRelatedRowsFromBrand(rows), nil
	default:
		return s.dbRO.GetQueries().GetRelatedProductsForSearch(ctx, params)
	}
}

func (s *ProductService) mapRelatedProductQueryRows(rows []relatedProductQueryRow) []models.RelatedProduct {
	related := make([]models.RelatedProduct, 0, len(rows))
	for _, row := range rows {
		if product, ok := s.toRelatedProduct(row); ok {
			related = append(related, product)
		}
	}
	return related
}

func (s *ProductService) toRelatedProduct(row relatedProductQueryRow) (models.RelatedProduct, bool) {
	if !row.Slug.Valid || row.Slug.String == "" {
		return models.RelatedProduct{}, false
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

	return models.RelatedProduct{
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
	}, true
}

func filterSearchRelatedRows(rows []queries.GetRelatedProductsForSearchRow) []queries.GetRelatedProductsForSearchRow {
	valid := make([]queries.GetRelatedProductsForSearchRow, 0, len(rows))
	for _, row := range rows {
		if strings.HasSuffix(row.ThumbnailPath, constants.EmptyImageFilename) {
			continue
		}
		if !row.Slug.Valid || row.Slug.String == "" {
			continue
		}
		valid = append(valid, row)
	}
	return valid
}

func toRelatedProductQueryRowsFromCategory(rows []queries.GetRelatedProductsByCategoryRow) []relatedProductQueryRow {
	result := make([]relatedProductQueryRow, len(rows))
	for i, row := range rows {
		result[i] = relatedProductQueryRow(row)
	}
	return result
}

func toRelatedProductQueryRowsFromParentCategory(rows []queries.GetRelatedProductsByParentCategoryRow) []relatedProductQueryRow {
	result := make([]relatedProductQueryRow, len(rows))
	for i, row := range rows {
		result[i] = relatedProductQueryRow(row)
	}
	return result
}

func toRelatedProductQueryRowsFromBrand(rows []queries.GetRelatedProductsByBrandRow) []relatedProductQueryRow {
	result := make([]relatedProductQueryRow, len(rows))
	for i, row := range rows {
		result[i] = relatedProductQueryRow(row)
	}
	return result
}

func toSearchRelatedRowsFromParentCategory(rows []queries.GetRelatedProductsForSearchByParentCategoryRow) []queries.GetRelatedProductsForSearchRow {
	result := make([]queries.GetRelatedProductsForSearchRow, len(rows))
	for i, row := range rows {
		result[i] = queries.GetRelatedProductsForSearchRow(row)
	}
	return result
}

func toSearchRelatedRowsFromBrand(rows []queries.GetRelatedProductsForSearchByBrandRow) []queries.GetRelatedProductsForSearchRow {
	result := make([]queries.GetRelatedProductsForSearchRow, len(rows))
	for i, row := range rows {
		result[i] = queries.GetRelatedProductsForSearchRow(row)
	}
	return result
}

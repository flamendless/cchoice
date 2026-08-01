package services

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/utils"
)

type BrandPageService struct {
	encoder encode.IEncode
	dbRO    database.IService
}

func NewBrandPageService(encoder encode.IEncode, dbRO database.IService) *BrandPageService {
	return &BrandPageService{
		encoder: encoder,
		dbRO:    dbRO,
	}
}

func (s *BrandPageService) GetBrandsListingPageData(
	ctx context.Context,
	getBrandLogoURL models.BrandLogoURLFunc,
) (*models.BrandsListingPageData, error) {
	rows, err := s.dbRO.GetQueries().GetBrandsForListingPage(ctx)
	if err != nil {
		return nil, errors.Join(errs.ErrBrand, err)
	}

	brands := make([]models.BrandListingCard, 0, len(rows))
	for _, row := range rows {
		slug := utils.NullStringValue(row.Slug)
		if slug == "" {
			slug = utils.BrandSlug(row.Name)
		}
		brands = append(brands, models.BrandListingCard{
			Slug:         slug,
			Name:         row.Name,
			LogoURL:      resolveBrandLogoURL(row.S3Url, utils.NullStringValue(row.Path), getBrandLogoURL),
			ProductCount: row.ProductCount,
			ComingSoon:   row.ProductCount == 0,
			PageURL:      utils.URLf("/brands/%s", slug),
		})
	}

	return &models.BrandsListingPageData{
		Brands: brands,
		SEO:    models.BrandsListingPageSEO(),
	}, nil
}

func (s *BrandPageService) GetBrandPageData(
	ctx context.Context,
	brandSlug string,
	getCDNURL models.CDNURLFunc,
	getBrandLogoURL models.BrandLogoURLFunc,
) (*models.BrandPageData, error) {
	brandSlug = strings.TrimSpace(brandSlug)
	if brandSlug == "" {
		return nil, errs.ErrNotFound
	}

	brand, err := s.dbRO.GetQueries().GetBrandBySlug(ctx, sql.NullString{String: brandSlug, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errors.Join(errs.ErrBrand, err)
	}

	bestSellingRows, err := s.dbRO.GetQueries().GetBestSellingProductsByBrandID(ctx, queries.GetBestSellingProductsByBrandIDParams{
		BrandID: brand.ID,
		Limit:   constants.DefaultBrandPageProductsLimit,
	})
	if err != nil {
		return nil, errors.Join(errs.ErrBrand, err)
	}

	highestDiscountRows, err := s.dbRO.GetQueries().GetHighestDiscountProductsByBrandID(ctx, queries.GetHighestDiscountProductsByBrandIDParams{
		BrandID: brand.ID,
		Limit:   constants.DefaultBrandPageProductsLimit,
	})
	if err != nil {
		return nil, errors.Join(errs.ErrBrand, err)
	}

	categoryRows, err := s.dbRO.GetQueries().GetBrandProductCategoriesForSections(ctx, brand.ID)
	if err != nil {
		return nil, errors.Join(errs.ErrBrand, err)
	}

	bestSellingProducts := bestSellingRowsToCategoryProducts(s.encoder, getCDNURL, bestSellingRows)
	highestDiscountProducts := highestDiscountRowsToCategoryProducts(s.encoder, getCDNURL, highestDiscountRows)

	groupedCategories := buildBrandGroupedCategorySections(brandSlug, s.encoder, categoryRows)

	return &models.BrandPageData{
		Slug:    brandSlug,
		Name:    brand.Name,
		LogoURL: resolveBrandLogoURL(brand.S3Url, utils.NullStringValue(brand.Path), getBrandLogoURL),
		BestSelling: buildBrandPrioritySection(
			"Best Selling",
			bestSellingProducts,
			utils.URLf("/brands/%s/sections/best-selling", brandSlug),
		),
		HighestDiscount: buildBrandPrioritySection(
			"Highest Discount",
			highestDiscountProducts,
			utils.URLf("/brands/%s/sections/highest-discount", brandSlug),
		),
		CategorySections: groupedCategories,
		SEO:              models.BrandPageSEO(brandSlug, brand.Name, resolveBrandLogoURL(brand.S3Url, utils.NullStringValue(brand.Path), getBrandLogoURL)),
	}, nil
}

func (s *BrandPageService) GetBrandPrioritySectionProducts(
	ctx context.Context,
	brandSlug string,
	section string,
	getCDNURL models.CDNURLFunc,
) ([]models.CategorySectionProduct, error) {
	brandSlug = strings.TrimSpace(brandSlug)
	if brandSlug == "" {
		return nil, errs.ErrNotFound
	}

	brand, err := s.dbRO.GetQueries().GetBrandBySlug(ctx, sql.NullString{String: brandSlug, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errors.Join(errs.ErrBrand, err)
	}

	switch strings.ToUpper(section) {
	case "BEST-SELLING":
		rows, err := s.dbRO.GetQueries().GetBestSellingProductsByBrandID(ctx, queries.GetBestSellingProductsByBrandIDParams{
			BrandID: brand.ID,
			Limit:   constants.DefaultBrandPageProductsLimit,
		})
		if err != nil {
			return nil, errors.Join(errs.ErrBrand, err)
		}
		return bestSellingRowsToCategoryProducts(s.encoder, getCDNURL, rows), nil
	case "HIGHEST-DISCOUNT":
		rows, err := s.dbRO.GetQueries().GetHighestDiscountProductsByBrandID(ctx, queries.GetHighestDiscountProductsByBrandIDParams{
			BrandID: brand.ID,
			Limit:   constants.DefaultBrandPageProductsLimit,
		})
		if err != nil {
			return nil, errors.Join(errs.ErrBrand, err)
		}
		return highestDiscountRowsToCategoryProducts(s.encoder, getCDNURL, rows), nil
	default:
		return nil, errs.ErrNotFound
	}
}

func (s *BrandPageService) GetBrandCategoryProducts(
	ctx context.Context,
	brandSlug string,
	categoryID string,
	getCDNURL models.CDNURLFunc,
) (models.CategorySectionProducts, error) {
	brandSlug = strings.TrimSpace(brandSlug)
	if brandSlug == "" {
		return models.CategorySectionProducts{}, errs.ErrNotFound
	}

	decodedCategoryID := s.encoder.Decode(categoryID)
	if decodedCategoryID == encode.INVALID {
		return models.CategorySectionProducts{}, errs.ErrInvalidParams
	}

	brand, err := s.dbRO.GetQueries().GetBrandBySlug(ctx, sql.NullString{String: brandSlug, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CategorySectionProducts{}, errs.ErrNotFound
		}
		return models.CategorySectionProducts{}, errors.Join(errs.ErrBrand, err)
	}

	category, err := s.dbRO.GetQueries().GetProductCategoryByID(ctx, decodedCategoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CategorySectionProducts{}, errs.ErrNotFound
		}
		return models.CategorySectionProducts{}, errors.Join(errs.ErrCategory, err)
	}

	products, err := s.dbRO.GetQueries().GetProductsByBrandAndCategoryID(ctx, queries.GetProductsByBrandAndCategoryIDParams{
		BrandID:    brand.ID,
		CategoryID: decodedCategoryID,
		Limit:      constants.DefaultLimitProducts,
	})
	if err != nil {
		return models.CategorySectionProducts{}, errors.Join(errs.ErrBrand, err)
	}

	return models.CategorySectionProducts{
		ID:          categoryID,
		Category:    utils.SlugToTile(category.Category.String),
		Subcategory: utils.SlugToTile(category.Subcategory.String),
		Products:    brandCategoryRowsToCategoryProducts(s.encoder, getCDNURL, products),
	}, nil
}

func (s *BrandPageService) ListBrandSitemapSlugs(ctx context.Context) ([]models.BrandSitemapSlug, error) {
	rows, err := s.dbRO.GetQueries().ListBrandSitemapSlugs(ctx)
	if err != nil {
		return nil, errors.Join(errs.ErrBrand, err)
	}

	result := make([]models.BrandSitemapSlug, 0, len(rows))
	for _, row := range rows {
		slug := utils.NullStringValue(row)
		if slug == "" {
			continue
		}
		result = append(result, models.BrandSitemapSlug{Slug: slug})
	}
	return result, nil
}

func (s *BrandPageService) ID() string {
	return "BrandPage"
}

func (s *BrandPageService) Log() {
	// no-op for IService
}

var _ IService = (*BrandPageService)(nil)

func resolveBrandLogoURL(s3URL sql.NullString, path string, getBrandLogoURL models.BrandLogoURLFunc) string {
	if s3URL.Valid && s3URL.String != "" {
		return s3URL.String
	}
	if path != "" {
		filename := filepath.Base(path)
		if filename != "" && filename != "." {
			return getBrandLogoURL(filename)
		}
	}
	return ""
}

func buildBrandPrioritySection(
	title string,
	products []models.CategorySectionProduct,
	showMoreURL string,
) models.BrandPrioritySection {
	totalCount := len(products)
	hasMore := totalCount > constants.DefaultBrandPagePriorityRowLimit
	displayProducts := products
	if hasMore {
		displayProducts = products[:constants.DefaultBrandPagePriorityRowLimit]
	}

	return models.BrandPrioritySection{
		Title:       title,
		Products:    displayProducts,
		TotalCount:  totalCount,
		ShowMoreURL: showMoreURL,
		HasMore:     hasMore,
	}
}

func buildBrandGroupedCategorySections(
	brandSlug string,
	encoder encode.IEncode,
	rows []queries.GetBrandProductCategoriesForSectionsRow,
) []models.BrandGroupedCategorySection {
	grouped := make(map[string]*models.BrandGroupedCategorySection)
	order := make([]string, 0)

	for _, row := range rows {
		if !row.Category.Valid || row.Category.String == "" {
			continue
		}
		categoryLabel := utils.SlugToTile(row.Category.String)
		subcategoryLabel := ""
		if row.Subcategory.Valid {
			subcategoryLabel = utils.SlugToTile(row.Subcategory.String)
		}

		if _, exists := grouped[categoryLabel]; !exists {
			grouped[categoryLabel] = &models.BrandGroupedCategorySection{
				Label:          categoryLabel,
				ScrollTargetID: utils.LabelToID(enums.MODULE_CATEGORY, categoryLabel),
				Subcategories:  []models.BrandSubcategorySection{},
			}
			order = append(order, categoryLabel)
		}

		categoryID := encoder.Encode(row.ID)
		grouped[categoryLabel].Subcategories = append(grouped[categoryLabel].Subcategories, models.BrandSubcategorySection{
			CategoryID:  categoryID,
			Label:       subcategoryLabel,
			ProductsURL: utils.URLf("/brands/%s/categories/%s/products", brandSlug, categoryID),
		})
	}

	result := make([]models.BrandGroupedCategorySection, 0, len(order))
	for _, label := range order {
		result = append(result, *grouped[label])
	}
	return result
}

func bestSellingRowsToCategoryProducts(
	encoder encode.IEncode,
	getCDNURL models.CDNURLFunc,
	rows []queries.GetBestSellingProductsByBrandIDRow,
) []models.CategorySectionProduct {
	res := make([]queries.GetProductsByCategoryIDRow, 0, len(rows))
	for _, row := range rows {
		res = append(res, queries.GetProductsByCategoryIDRow{
			ID:                       row.ID,
			Serial:                   row.Serial,
			Slug:                     row.Slug,
			Name:                     row.Name,
			Description:              row.Description,
			UnitPriceWithVat:         row.UnitPriceWithVat,
			UnitPriceWithVatCurrency: row.UnitPriceWithVatCurrency,
			SalePriceWithVat:         row.SalePriceWithVat,
			SalePriceWithVatCurrency: row.SalePriceWithVatCurrency,
			IsOnSale:                 row.IsOnSale,
			DiscountType:             row.DiscountType,
			DiscountValue:            row.DiscountValue,
			BrandName:                row.BrandName,
			ThumbnailPath:            row.ThumbnailPath,
			CdnUrl:                   row.CdnUrl,
			CdnUrlThumbnail:          row.CdnUrlThumbnail,
		})
	}
	return models.ToCategorySectionProducts(encoder, getCDNURL, res)
}

func highestDiscountRowsToCategoryProducts(
	encoder encode.IEncode,
	getCDNURL models.CDNURLFunc,
	rows []queries.GetHighestDiscountProductsByBrandIDRow,
) []models.CategorySectionProduct {
	res := make([]queries.GetProductsByCategoryIDRow, 0, len(rows))
	for _, row := range rows {
		res = append(res, queries.GetProductsByCategoryIDRow{
			ID:                       row.ID,
			Serial:                   row.Serial,
			Slug:                     row.Slug,
			Name:                     row.Name,
			Description:              row.Description,
			UnitPriceWithVat:         row.UnitPriceWithVat,
			UnitPriceWithVatCurrency: row.UnitPriceWithVatCurrency,
			SalePriceWithVat:         sql.NullInt64{Int64: row.SalePriceWithVat, Valid: true},
			SalePriceWithVatCurrency: sql.NullString{String: row.SalePriceWithVatCurrency, Valid: true},
			IsOnSale:                 row.IsOnSale,
			DiscountType:             sql.NullString{String: row.DiscountType, Valid: row.DiscountType != ""},
			DiscountValue:            sql.NullInt64{Int64: row.DiscountValue, Valid: true},
			BrandName:                row.BrandName,
			ThumbnailPath:            row.ThumbnailPath,
			CdnUrl:                   row.CdnUrl,
			CdnUrlThumbnail:          row.CdnUrlThumbnail,
		})
	}
	return models.ToCategorySectionProducts(encoder, getCDNURL, res)
}

func brandCategoryRowsToCategoryProducts(
	encoder encode.IEncode,
	getCDNURL models.CDNURLFunc,
	rows []queries.GetProductsByBrandAndCategoryIDRow,
) []models.CategorySectionProduct {
	res := make([]queries.GetProductsByCategoryIDRow, 0, len(rows))
	for _, row := range rows {
		res = append(res, queries.GetProductsByCategoryIDRow(row))
	}
	return models.ToCategorySectionProducts(encoder, getCDNURL, res)
}

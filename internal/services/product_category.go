package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

type ProductCategoryService struct {
	encoder  encode.IEncode
	dbRO     database.IService
	dbRW     database.IService
	staffLog *StaffLogsService
}

func NewProductCategoryService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
) *ProductCategoryService {
	if staffLog == nil {
		panic("StaffLogsService is required")
	}
	return &ProductCategoryService{
		encoder:  encoder,
		dbRO:     dbRO,
		dbRW:     dbRW,
		staffLog: staffLog,
	}
}

type CreateCategoriesParams struct {
	Mode          string
	CategoryName  string
	Subcategories []string
}

func (s *ProductCategoryService) GetCategoriesListingPaginated(
	ctx context.Context,
	search string,
	page, perPage int,
) ([]models.AdminCategoryListItem, int64, int, error) {
	searchParam := sql.NullString{String: search, Valid: search != ""}

	totalCount, err := s.dbRO.GetQueries().CountDistinctCategoriesForAdmin(ctx, searchParam)
	if err != nil {
		return nil, 0, 0, errors.Join(errs.ErrCategory, err)
	}

	page = models.ClampPage(page, totalCount, perPage)
	offset := int64((page - 1) * perPage)

	rows, err := s.dbRO.GetQueries().GetDistinctCategoriesForAdminPaginated(ctx, queries.GetDistinctCategoriesForAdminPaginatedParams{
		Search: searchParam,
		Limit:  int64(perPage),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, 0, errors.Join(errs.ErrCategory, err)
	}

	items := make([]models.AdminCategoryListItem, 0, len(rows))
	for _, row := range rows {
		if !row.Category.Valid {
			continue
		}
		items = append(items, models.AdminCategoryListItem{
			Category:           row.Category.String,
			SubcategoriesCount: row.SubcategoriesCount,
			ProductsCount:      row.ProductsCount,
		})
	}

	return items, totalCount, page, nil
}

func (s *ProductCategoryService) GetSubcategoriesForCategory(
	ctx context.Context,
	categoryName string,
) ([]models.AdminSubcategoryRow, error) {
	rows, err := s.dbRO.GetQueries().GetSubcategoriesByCategoryForAdmin(ctx, sql.NullString{
		String: categoryName,
		Valid:  categoryName != "",
	})
	if err != nil {
		return nil, errors.Join(errs.ErrCategory, err)
	}

	result := make([]models.AdminSubcategoryRow, 0, len(rows))
	for _, row := range rows {
		subcategory := ""
		if row.Subcategory.Valid {
			subcategory = row.Subcategory.String
		}
		promoted := row.PromotedAtHomepage.Valid && row.PromotedAtHomepage.Bool
		result = append(result, models.AdminSubcategoryRow{
			ID:          s.encoder.Encode(row.ID),
			Subcategory: subcategory,
			Promoted:    promoted,
		})
	}

	return result, nil
}

func (s *ProductCategoryService) GetAllCategoryNames(ctx context.Context) ([]string, error) {
	rows, err := s.dbRO.GetQueries().GetAllCategoryNames(ctx)
	if err != nil {
		return nil, errors.Join(errs.ErrCategory, err)
	}

	names := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Valid {
			names = append(names, row.String)
		}
	}
	return names, nil
}

func (s *ProductCategoryService) GetCategoryPageData(
	ctx context.Context,
	categorySlug string,
	subcategorySlug string,
	getCDNURL models.CDNURLFunc,
) (*models.CategoryPageData, error) {
	categorySlug = strings.TrimSpace(categorySlug)
	if categorySlug == "" {
		return nil, errs.ErrNotFound
	}

	subcategorySlug = strings.TrimSpace(subcategorySlug)
	var products []queries.GetProductsByCategoryIDRow

	if subcategorySlug != "" {
		_, err := s.dbRO.GetQueries().GetProductCategoryByCategoryAndSubcategory(ctx, queries.GetProductCategoryByCategoryAndSubcategoryParams{
			Category:    sql.NullString{String: categorySlug, Valid: true},
			Subcategory: sql.NullString{String: subcategorySlug, Valid: true},
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errs.ErrNotFound
			}
			return nil, errors.Join(errs.ErrCategory, err)
		}

		rows, err := s.dbRO.GetQueries().GetProductsByCategoryAndSubcategorySlug(ctx, queries.GetProductsByCategoryAndSubcategorySlugParams{
			Category:    sql.NullString{String: categorySlug, Valid: true},
			Subcategory: sql.NullString{String: subcategorySlug, Valid: true},
			Limit:       constants.DefaultCategoryPageProductLimit,
		})
		if err != nil {
			return nil, errors.Join(errs.ErrCategory, err)
		}
		products = subcategorySlugRowsToIDRows(rows)
	} else {
		exists, err := s.dbRO.GetQueries().CategoryNameExists(ctx, sql.NullString{
			String: categorySlug,
			Valid:  true,
		})
		if err != nil {
			return nil, errors.Join(errs.ErrCategory, err)
		}
		if !exists {
			return nil, errs.ErrNotFound
		}

		rows, err := s.dbRO.GetQueries().GetProductsByCategorySlug(ctx, queries.GetProductsByCategorySlugParams{
			Category: sql.NullString{String: categorySlug, Valid: true},
			Limit:    constants.DefaultCategoryPageProductLimit,
		})
		if err != nil {
			return nil, errors.Join(errs.ErrCategory, err)
		}
		products = categorySlugRowsToIDRows(rows)
	}

	if len(products) == 0 {
		return nil, errs.ErrNotFound
	}

	categoryLabel := utils.SlugToTile(categorySlug)
	subcategoryLabel := ""
	if subcategorySlug != "" {
		subcategoryLabel = utils.SlugToTile(subcategorySlug)
	}

	return &models.CategoryPageData{
		CategorySlug:     categorySlug,
		SubcategorySlug:  subcategorySlug,
		CategoryLabel:    categoryLabel,
		SubcategoryLabel: subcategoryLabel,
		Products:         models.ToCategorySectionProducts(s.encoder, getCDNURL, products),
		SEO:              models.CategoryPageSEO(categorySlug, subcategorySlug),
	}, nil
}

func (s *ProductCategoryService) ListCategorySitemapSlugs(ctx context.Context) ([]models.CategorySitemapSlug, error) {
	rows, err := s.dbRO.GetQueries().ListCategorySitemapEntries(ctx)
	if err != nil {
		return nil, errors.Join(errs.ErrCategory, err)
	}

	result := make([]models.CategorySitemapSlug, 0, len(rows))
	for _, row := range rows {
		if !row.Category.Valid || row.Category.String == "" {
			continue
		}
		if !row.Subcategory.Valid || row.Subcategory.String == "" {
			continue
		}
		result = append(result, models.CategorySitemapSlug{
			Category:    row.Category.String,
			Subcategory: row.Subcategory.String,
		})
	}
	return result, nil
}

func categorySlugRowsToIDRows(rows []queries.GetProductsByCategorySlugRow) []queries.GetProductsByCategoryIDRow {
	res := make([]queries.GetProductsByCategoryIDRow, 0, len(rows))
	for _, row := range rows {
		res = append(res, queries.GetProductsByCategoryIDRow(row))
	}
	return res
}

func subcategorySlugRowsToIDRows(rows []queries.GetProductsByCategoryAndSubcategorySlugRow) []queries.GetProductsByCategoryIDRow {
	res := make([]queries.GetProductsByCategoryIDRow, 0, len(rows))
	for _, row := range rows {
		res = append(res, queries.GetProductsByCategoryIDRow(row))
	}
	return res
}

func (s *ProductCategoryService) CreateCategories(
	ctx context.Context,
	staffID string,
	params CreateCategoriesParams,
) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionCreate,
			constants.ModuleCategories,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("[ProductCategoryService] create log", zap.Error(err))
		}
	}()

	subcategories := normalizeSubcategoryNames(params.Subcategories)
	if len(subcategories) == 0 {
		result = "subcategories required"
		return errs.ErrCategory
	}

	categoryName := strings.TrimSpace(params.CategoryName)
	if categoryName == "" {
		result = "category name required"
		return errs.ErrCategory
	}

	switch params.Mode {
	case "new":
		exists, err := s.dbRO.GetQueries().CategoryNameExists(ctx, sql.NullString{
			String: categoryName,
			Valid:  true,
		})
		if err != nil {
			result = err.Error()
			return errors.Join(errs.ErrCategory, err)
		}
		if exists {
			result = errs.ErrCategoryAlreadyExists.Error()
			return errs.ErrCategoryAlreadyExists
		}
	case "existing":
		exists, err := s.dbRO.GetQueries().CategoryNameExists(ctx, sql.NullString{
			String: categoryName,
			Valid:  true,
		})
		if err != nil {
			result = err.Error()
			return errors.Join(errs.ErrCategory, err)
		}
		if !exists {
			result = errs.ErrCategoryNotFound.Error()
			return errs.ErrCategoryNotFound
		}
	default:
		result = "invalid mode"
		return errs.ErrCategory
	}

	for _, subcategory := range subcategories {
		_, err := s.dbRO.GetQueries().GetProductCategoryByCategoryAndSubcategory(ctx, queries.GetProductCategoryByCategoryAndSubcategoryParams{
			Category:    sql.NullString{String: categoryName, Valid: true},
			Subcategory: sql.NullString{String: subcategory, Valid: true},
		})
		if err == nil {
			result = errs.ErrCategoryPairExists.Error()
			return errs.ErrCategoryPairExists
		}
		if !errors.Is(err, sql.ErrNoRows) {
			result = err.Error()
			return errors.Join(errs.ErrCategory, err)
		}

		if _, err := s.dbRW.GetQueries().CreateProductCategory(ctx, queries.CreateProductCategoryParams{
			Category:    sql.NullString{String: categoryName, Valid: true},
			Subcategory: sql.NullString{String: subcategory, Valid: true},
		}); err != nil {
			result = err.Error()
			return errors.Join(errs.ErrCategory, err)
		}
	}

	result = fmt.Sprintf("success. category '%s', subcategories %d", categoryName, len(subcategories))
	return nil
}

func normalizeSubcategoryNames(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	result := make([]string, 0, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func (s *ProductCategoryService) ID() string {
	return "ProductCategory"
}

func (s *ProductCategoryService) Log() {
	logs.Log().Info("[ProductCategoryService] Loaded")
}

var _ IService = (*ProductCategoryService)(nil)

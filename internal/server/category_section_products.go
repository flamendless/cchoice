package server

import (
	"context"
	"strings"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/errs"
	"cchoice/internal/utils"
)

func (s *Server) loadCategorySectionProducts(
	ctx context.Context,
	categoryID string,
	brandID int64,
) (models.CategorySectionProducts, error) {
	decodedCategoryID := s.encoder.Decode(categoryID)
	if decodedCategoryID == encode.INVALID {
		return models.CategorySectionProducts{}, errs.ErrInvalidParams
	}

	category, err := s.dbRO.GetQueries().GetProductCategoryByID(ctx, decodedCategoryID)
	if err != nil {
		return models.CategorySectionProducts{}, err
	}

	if category.Category.String == "" {
		return models.CategorySectionProducts{
			ID:          categoryID,
			Category:    "",
			Subcategory: "",
			Products:    nil,
		}, nil
	}

	products, err := s.dbRO.GetQueries().GetProductsByCategoryID(ctx, queries.GetProductsByCategoryIDParams{
		CategoryID: decodedCategoryID,
		BrandID:    brandID,
		Limit:      constants.DefaultLimitProducts,
	})
	if err != nil {
		return models.CategorySectionProducts{}, err
	}

	productsWithValidImages := filterProductsWithValidImages(products)

	return models.CategorySectionProducts{
		ID:          categoryID,
		Category:    utils.SlugToTile(category.Category.String),
		Subcategory: utils.SlugToTile(category.Subcategory.String),
		Products:    models.ToCategorySectionProducts(s.encoder, s.GetCDNURL, productsWithValidImages),
	}, nil
}

func filterProductsWithValidImages(products []queries.GetProductsByCategoryIDRow) []queries.GetProductsByCategoryIDRow {
	validProducts := make([]int, 0, len(products))
	for i, product := range products {
		if !strings.HasSuffix(product.ThumbnailPath, constants.EmptyImageFilename) {
			validProducts = append(validProducts, i)
		}
	}

	productsWithValidImages := make([]queries.GetProductsByCategoryIDRow, 0, len(validProducts))
	for _, i := range validProducts {
		productsWithValidImages = append(productsWithValidImages, products[i])
	}
	return productsWithValidImages
}

func firstSubcategoryIDs(sections []models.GroupedCategorySection, limit int) []string {
	if limit <= 0 {
		return nil
	}

	ids := make([]string, 0, limit)
	for _, category := range sections {
		for _, subcategory := range category.Subcategories {
			ids = append(ids, subcategory.CategoryID)
			if len(ids) >= limit {
				return ids
			}
		}
	}
	return ids
}

func firstBrandSubcategoryIDs(sections []models.BrandGroupedCategorySection, limit int) []string {
	if limit <= 0 {
		return nil
	}

	ids := make([]string, 0, limit)
	for _, category := range sections {
		for _, subcategory := range category.Subcategories {
			ids = append(ids, subcategory.CategoryID)
			if len(ids) >= limit {
				return ids
			}
		}
	}
	return ids
}

func (s *Server) preloadCategorySectionProducts(
	ctx context.Context,
	sections []models.GroupedCategorySection,
	brandID int64,
	limit int,
) map[string]models.CategorySectionProducts {
	ids := firstSubcategoryIDs(sections, limit)
	result := make(map[string]models.CategorySectionProducts, len(ids))
	for _, id := range ids {
		sectionProducts, err := s.loadCategorySectionProducts(ctx, id, brandID)
		if err != nil {
			continue
		}
		result[id] = sectionProducts
	}
	return result
}

func (s *Server) preloadBrandCategorySectionProducts(
	ctx context.Context,
	brandSlug string,
	sections []models.BrandGroupedCategorySection,
	limit int,
) map[string]models.CategorySectionProducts {
	ids := firstBrandSubcategoryIDs(sections, limit)
	result := make(map[string]models.CategorySectionProducts, len(ids))
	for _, id := range ids {
		sectionProducts, err := s.services.brandPage.GetBrandCategoryProducts(ctx, brandSlug, id, s.GetCDNURL)
		if err != nil {
			continue
		}
		result[id] = sectionProducts
	}
	return result
}

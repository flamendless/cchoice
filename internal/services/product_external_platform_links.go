package services

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"cchoice/cmd/web/models"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/utils"
)

func validateExternalPlatformLinks(links []ExternalPlatformLinkInput) error {
	seen := make(map[string]struct{}, len(links))

	for _, link := range links {
		platform := strings.TrimSpace(link.Platform)
		url := strings.TrimSpace(link.URL)

		if platform == "" && url == "" {
			continue
		}

		if platform == "" || url == "" {
			return errs.ErrInvalidExternalPlatformLink
		}

		if enums.ParseExternalPlatformToEnum(platform) == enums.EXTERNAL_PLATFORM_UNDEFINED {
			return errs.ErrInvalidExternalPlatformLink
		}

		if err := utils.ValidateExternalURL(url); err != nil {
			return errs.ErrInvalidExternalPlatformLink
		}

		platformKey := strings.ToUpper(platform)
		if _, exists := seen[platformKey]; exists {
			return errs.ErrInvalidExternalPlatformLink
		}
		seen[platformKey] = struct{}{}
	}

	return nil
}

func (s *ProductService) SyncExternalPlatformLinks(
	ctx context.Context,
	productID int64,
	links []ExternalPlatformLinkInput,
) error {
	if err := validateExternalPlatformLinks(links); err != nil {
		return err
	}

	if err := s.dbRW.GetQueries().DeleteProductExternalPlatformLinksByProductID(ctx, productID); err != nil {
		return fmt.Errorf("failed to delete external platform links: %w", err)
	}

	for _, link := range links {
		platform := strings.TrimSpace(link.Platform)
		url := strings.TrimSpace(link.URL)
		if platform == "" || url == "" {
			continue
		}

		if _, err := s.dbRW.GetQueries().CreateProductExternalPlatformLink(ctx, queries.CreateProductExternalPlatformLinkParams{
			ProductID: productID,
			Platform:  strings.ToUpper(platform),
			Url:       url,
		}); err != nil {
			return fmt.Errorf("failed to create external platform link: %w", err)
		}
	}

	return nil
}

func (s *ProductService) getExternalPlatformLinksForProduct(
	ctx context.Context,
	productID int64,
) ([]ExternalPlatformLinkInput, error) {
	rows, err := s.dbRO.GetQueries().GetProductExternalPlatformLinksByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get external platform links: %w", err)
	}

	links := make([]ExternalPlatformLinkInput, 0, len(rows))
	for _, row := range rows {
		links = append(links, ExternalPlatformLinkInput{
			Platform: row.Platform,
			URL:      row.Url,
		})
	}

	return links, nil
}

func (s *ProductService) buildProductExternalPlatformLinks(
	ctx context.Context,
	productID int64,
) ([]models.ProductExternalPlatformLink, error) {
	rawLinks, err := s.getExternalPlatformLinksForProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	links := make([]models.ProductExternalPlatformLink, 0, len(rawLinks))
	for _, link := range rawLinks {
		platform := enums.ParseExternalPlatformToEnum(link.Platform)
		if platform == enums.EXTERNAL_PLATFORM_UNDEFINED {
			continue
		}

		links = append(links, models.ProductExternalPlatformLink{
			Platform: platform,
			URL:      link.URL,
		})
	}

	sortExternalPlatformLinks(links)

	return links, nil
}

func sortExternalPlatformLinks(links []models.ProductExternalPlatformLink) {
	order := make(map[enums.ExternalPlatform]int, len(enums.AllExternalPlatforms))
	for i, platform := range enums.AllExternalPlatforms {
		order[platform] = i
	}

	slices.SortFunc(links, func(a, b models.ProductExternalPlatformLink) int {
		return order[a.Platform] - order[b.Platform]
	})
}

package seo

import (
	"encoding/json"
	"fmt"
	"strings"

	"cchoice/internal/utils"
)

func CategoryPageURL(siteBaseURL, categorySlug, subcategorySlug string) string {
	base := strings.TrimSuffix(siteBaseURL, "/")
	if subcategorySlug != "" && !strings.EqualFold(subcategorySlug, categorySlug) {
		return fmt.Sprintf("%s/categories/%s/%s", base, categorySlug, subcategorySlug)
	}
	return fmt.Sprintf("%s/categories/%s", base, categorySlug)
}

type CategoryMeta struct {
	Title          string
	Description    string
	CanonicalURL   string
	OGImage        string
	OGType         string
	Robots         string
	Keywords       string
	TwitterCard    string
	StructuredData json.RawMessage
}

func GenerateCategoryMeta(categorySlug, subcategorySlug string) CategoryMeta {
	categoryLabel := utils.SlugToTile(categorySlug)
	subcategoryLabel := utils.SlugToTile(subcategorySlug)

	title, description := buildCategoryMetaText(categorySlug, subcategorySlug, categoryLabel, subcategoryLabel)

	canonicalURL := CategoryPageURL(utils.SiteURL("/"), categorySlug, subcategorySlug)
	keywords := strings.Join([]string{
		categoryLabel,
		subcategoryLabel,
		"cchoice",
		"c-choice",
		"power tools",
		"construction supplies",
		"philippines",
	}, ", ")

	return CategoryMeta{
		Title:          title,
		Description:    description,
		CanonicalURL:   canonicalURL,
		OGImage:        DefaultOGImage,
		OGType:         "website",
		Robots:         ProductRobots,
		Keywords:       keywords,
		TwitterCard:    ProductTwitterCard,
		StructuredData: BuildCategoryStructuredData(categorySlug, subcategorySlug, canonicalURL, title, description),
	}
}

func buildCategoryMetaText(categorySlug, subcategorySlug, categoryLabel, subcategoryLabel string) (string, string) {
	if subcategorySlug != "" && !strings.EqualFold(subcategorySlug, categorySlug) {
		return fmt.Sprintf("%s - %s | C-Choice Construction Supply", subcategoryLabel, categoryLabel),
			fmt.Sprintf(
				"Shop %s %s power tools and construction supplies at C-Choice Philippines. Browse quality products from trusted brands with competitive pricing.",
				categoryLabel,
				subcategoryLabel,
			)
	}

	return categoryLabel + " | C-Choice Construction Supply",
		fmt.Sprintf(
			"Shop %s power tools and construction supplies at C-Choice Philippines. Browse quality products from trusted brands with competitive pricing.",
			categoryLabel,
		)
}

func buildCategoryBreadcrumbItems(categorySlug, subcategorySlug, canonicalURL string) []breadcrumbItem {
	homeURL := strings.TrimSuffix(utils.SiteURL("/"), "/") + "/"
	items := []breadcrumbItem{
		{Type: "ListItem", Position: 1, Name: "Home", Item: homeURL},
	}

	position := 2
	categoryURL := CategoryPageURL(utils.SiteURL("/"), categorySlug, "")
	items = append(items, breadcrumbItem{
		Type:     "ListItem",
		Position: position,
		Name:     utils.SlugToTile(categorySlug),
		Item:     categoryURL,
	})
	position++

	if subcategorySlug != "" && !strings.EqualFold(subcategorySlug, categorySlug) {
		items = append(items, breadcrumbItem{
			Type:     "ListItem",
			Position: position,
			Name:     utils.SlugToTile(subcategorySlug),
			Item:     canonicalURL,
		})
	}

	return items
}

func BuildCategoryStructuredData(categorySlug, subcategorySlug, canonicalURL, title, description string) json.RawMessage {
	type breadcrumbList struct {
		Type     string           `json:"@type"`
		ItemList []breadcrumbItem `json:"itemListElement"`
	}
	type collectionPage struct {
		Type        string `json:"@type"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		URL         string `json:"url"`
	}
	type graph struct {
		Context string `json:"@context"`
		Graph   []any  `json:"@graph"`
	}

	metaTitle := title
	metaDescription := description
	if metaTitle == "" || metaDescription == "" {
		categoryLabel := utils.SlugToTile(categorySlug)
		subcategoryLabel := utils.SlugToTile(subcategorySlug)
		metaTitle, metaDescription = buildCategoryMetaText(categorySlug, subcategorySlug, categoryLabel, subcategoryLabel)
	}

	payload := graph{
		Context: "https://schema.org",
		Graph: []any{
			collectionPage{
				Type:        "CollectionPage",
				Name:        metaTitle,
				Description: metaDescription,
				URL:         canonicalURL,
			},
			breadcrumbList{
				Type:     "BreadcrumbList",
				ItemList: buildCategoryBreadcrumbItems(categorySlug, subcategorySlug, canonicalURL),
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return data
}

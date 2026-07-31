package seo

import (
	"encoding/json"
	"fmt"
	"strings"

	"cchoice/internal/utils"
)

func BrandsListingPageURL(siteBaseURL string) string {
	return strings.TrimSuffix(siteBaseURL, "/") + "/brands/"
}

func BrandPageURL(siteBaseURL, brandSlug string) string {
	return fmt.Sprintf("%s/brands/%s", strings.TrimSuffix(siteBaseURL, "/"), brandSlug)
}

type BrandMeta struct {
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

func GenerateBrandsListingMeta() BrandMeta {
	title := "Brands | C-Choice Construction Supply"
	description := "Browse all power tool and construction supply brands at C-Choice Philippines. Shop quality products from trusted manufacturers."
	canonicalURL := BrandsListingPageURL(utils.SiteURL("/"))

	return BrandMeta{
		Title:          title,
		Description:    description,
		CanonicalURL:   canonicalURL,
		OGImage:        DefaultOGImage,
		OGType:         "website",
		Robots:         ProductRobots,
		Keywords:       "brands, power tools, construction supplies, cchoice, philippines",
		TwitterCard:    ProductTwitterCard,
		StructuredData: BuildBrandsListingStructuredData(canonicalURL, title, description),
	}
}

func GenerateBrandMeta(brandSlug, brandName, ogImageURL string) BrandMeta {
	title := fmt.Sprintf("%s | C-Choice Construction Supply", brandName)
	description := fmt.Sprintf(
		"Shop %s power tools and construction supplies at C-Choice Philippines. Browse best sellers, deals, and products from %s.",
		brandName,
		brandName,
	)
	canonicalURL := BrandPageURL(utils.SiteURL("/"), brandSlug)
	keywords := strings.Join([]string{
		brandName,
		"cchoice",
		"c-choice",
		"power tools",
		"construction supplies",
		"philippines",
	}, ", ")

	ogImage := strings.TrimSpace(ogImageURL)
	if ogImage == "" {
		ogImage = DefaultOGImage
	}

	return BrandMeta{
		Title:          title,
		Description:    description,
		CanonicalURL:   canonicalURL,
		OGImage:        ogImage,
		OGType:         "website",
		Robots:         ProductRobots,
		Keywords:       keywords,
		TwitterCard:    ProductTwitterCard,
		StructuredData: BuildBrandStructuredData(brandSlug, brandName, canonicalURL, title, description),
	}
}

func BuildBrandsListingStructuredData(canonicalURL, title, description string) json.RawMessage {
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

	payload := graph{
		Context: "https://schema.org",
		Graph: []any{
			collectionPage{
				Type:        "CollectionPage",
				Name:        title,
				Description: description,
				URL:         canonicalURL,
			},
			breadcrumbList{
				Type:     "BreadcrumbList",
				ItemList: buildBrandsListingBreadcrumbItems(canonicalURL),
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return data
}

func buildBrandsListingBreadcrumbItems(canonicalURL string) []breadcrumbItem {
	homeURL := siteHomeURLFromCanonical(canonicalURL)

	return []breadcrumbItem{
		{Type: "ListItem", Position: 1, Name: "Home", Item: homeURL},
		{Type: "ListItem", Position: 2, Name: "Brands", Item: canonicalURL},
	}
}

func buildBrandBreadcrumbItems(brandSlug, brandName, canonicalURL string) []breadcrumbItem {
	homeURL := siteHomeURLFromCanonical(canonicalURL)
	brandsURL := BrandsListingPageURL(strings.TrimSuffix(homeURL, "/"))

	return []breadcrumbItem{
		{Type: "ListItem", Position: 1, Name: "Home", Item: homeURL},
		{Type: "ListItem", Position: 2, Name: "Brands", Item: brandsURL},
		{Type: "ListItem", Position: 3, Name: brandName, Item: canonicalURL},
	}
}

func siteHomeURLFromCanonical(canonicalURL string) string {
	if idx := strings.Index(canonicalURL, "/brands"); idx > 0 {
		return canonicalURL[:idx] + "/"
	}

	base := strings.TrimSuffix(canonicalURL, "/")
	if strings.HasSuffix(base, "/brands") {
		return base[:len(base)-len("/brands")] + "/"
	}

	return base + "/"
}

func BuildBrandStructuredData(brandSlug, brandName, canonicalURL, title, description string) json.RawMessage {
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
		metaTitle = fmt.Sprintf("%s | C-Choice Construction Supply", brandName)
		metaDescription = fmt.Sprintf(
			"Shop %s power tools and construction supplies at C-Choice Philippines. Browse best sellers, deals, and products from %s.",
			brandName,
			brandName,
		)
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
				ItemList: buildBrandBreadcrumbItems(brandSlug, brandName, canonicalURL),
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return data
}

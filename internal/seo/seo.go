package seo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cchoice/internal/enums"
	"cchoice/internal/utils"
)

const (
	ProductRobots      = "index, follow, max-image-preview:large"
	ProductOGType      = "product"
	ProductTwitterCard = "summary_large_image"
	DefaultOGImage     = "https://imagedelivery.net/YnES7emCTPeSEVA2N0dB_g/favicons-192x192/public"
	SitemapPlaceholder = "# __SITEMAP__"
)

type Product struct {
	BrandName          string
	Name               string
	Serial             string
	Description        string
	ProductCategory    string
	ProductSubcategory string
	OnSale             bool
}

type ProductMeta struct {
	Title          string
	Description    string
	CanonicalURL   string
	OGImage        string
	OGType         string
	Robots         string
	Keywords       string
	TwitterCard    string
	PriceAmount    string
	PriceCurrency  string
	StructuredData json.RawMessage
}

type SitemapEntry struct {
	Loc        string
	LastMod    time.Time
	ChangeFreq string
	Priority   string
}

type breadcrumbItem struct {
	Type     string `json:"@type"`
	Position int    `json:"position"`
	Name     string `json:"name"`
	Item     string `json:"item,omitempty"`
}

func ProductCanonicalURL(siteBaseURL, slug string) string {
	return strings.TrimSuffix(siteBaseURL, "/") + "/product/" + slug
}

func InjectSitemapLine(content, sitemapURL string) string {
	return strings.Replace(content, SitemapPlaceholder, "Sitemap: "+sitemapURL, 1)
}

func EscapeXML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"'", "&apos;",
		"\"", "&quot;",
	)
	return replacer.Replace(s)
}

func WriteSitemapURL(buf *bytes.Buffer, entry SitemapEntry) {
	fmt.Fprintf(buf, "<url><loc>%s</loc>", EscapeXML(entry.Loc))
	fmt.Fprintf(buf, "<lastmod>%s</lastmod>", entry.LastMod.Format("2006-01-02"))
	fmt.Fprintf(buf, "<changefreq>%s</changefreq>", entry.ChangeFreq)
	fmt.Fprintf(buf, "<priority>%s</priority></url>", entry.Priority)
}

func BuildSitemapXML(homeURL string, products []SitemapEntry) string {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)

	WriteSitemapURL(&buf, SitemapEntry{
		Loc:        homeURL,
		LastMod:    time.Now().UTC(),
		ChangeFreq: "daily",
		Priority:   "1.0",
	})

	for _, product := range products {
		WriteSitemapURL(&buf, product)
	}

	buf.WriteString(`</urlset>`)
	return buf.String()
}

func buildProductTitle(product Product) string {
	const saleTitlePrefix = "SALE! "
	parts := []string{product.BrandName, product.Name}
	if product.ProductCategory != "" {
		parts = append(parts, utils.SlugToTile(product.ProductCategory))
	}
	if product.ProductSubcategory != "" {
		parts = append(parts, utils.SlugToTile(product.ProductSubcategory))
	}
	parts = append(parts, "Power Tools")
	title := strings.Join(parts, " ") + " | Price, Specs, Buy Online | C-Choice"
	if product.OnSale {
		title = saleTitlePrefix + title
	}
	return title
}

func GenerateProductMeta(
	product Product,
	canonicalURL string,
	homeURL string,
	imageURL string,
	priceAmount string,
	priceCurrency string,
) ProductMeta {
	title := buildProductTitle(product)

	description := strings.TrimSpace(product.Description)
	if description == "" {
		description = fmt.Sprintf(
			"Shop %s %s (%s) at C-Choice. Power tools and construction supplies in the Philippines with competitive pricing.",
			product.BrandName,
			product.Name,
			product.Serial,
		)
	} else if len(description) > 155 {
		description = description[:152] + "..."
	}

	keywordsParts := []string{
		product.BrandName,
		product.Name,
		product.Serial,
		product.ProductCategory,
		product.ProductSubcategory,
	}
	keywordsParts = append(keywordsParts, BaseProductKeywords()...)
	if strings.EqualFold(product.BrandName, "Bosch") {
		keywordsParts = append(keywordsParts, "bosch philippines", "bosh")
	}
	if strings.EqualFold(product.BrandName, "INGCO") {
		keywordsParts = append(keywordsParts, "ingco philippines")
	}
	keywords := strings.Join(keywordsParts, ", ")

	ogImage := imageURL
	if ogImage == "" {
		ogImage = DefaultOGImage
	}

	return ProductMeta{
		Title:          title,
		Description:    description,
		CanonicalURL:   canonicalURL,
		OGImage:        ogImage,
		OGType:         ProductOGType,
		Robots:         ProductRobots,
		Keywords:       keywords,
		TwitterCard:    ProductTwitterCard,
		PriceAmount:    priceAmount,
		PriceCurrency:  priceCurrency,
		StructuredData: BuildProductStructuredData(product, canonicalURL, ogImage, priceAmount, priceCurrency, homeURL),
	}
}

func categoryBreadcrumbURL(siteBaseURL, categorySlug string) string {
	label := utils.SlugToTile(categorySlug)
	anchor := utils.LabelToID(enums.MODULE_CATEGORY, label)
	return strings.TrimSuffix(siteBaseURL, "/") + "/#" + anchor
}

func buildProductBreadcrumbItems(product Product, siteBaseURL string) []breadcrumbItem {
	homeURL := strings.TrimSuffix(siteBaseURL, "/") + "/"
	items := []breadcrumbItem{
		{Type: "ListItem", Position: 1, Name: "Home", Item: homeURL},
	}

	position := 2
	var categoryURL string
	if product.ProductCategory != "" {
		categoryURL = categoryBreadcrumbURL(siteBaseURL, product.ProductCategory)
		items = append(items, breadcrumbItem{
			Type:     "ListItem",
			Position: position,
			Name:     utils.SlugToTile(product.ProductCategory),
			Item:     categoryURL,
		})
		position++
	}

	subcategory := product.ProductSubcategory
	if subcategory != "" && !strings.EqualFold(subcategory, product.ProductCategory) {
		subcategoryURL := categoryURL
		if subcategoryURL == "" {
			subcategoryURL = homeURL
		}
		items = append(items, breadcrumbItem{
			Type:     "ListItem",
			Position: position,
			Name:     utils.SlugToTile(subcategory),
			Item:     subcategoryURL,
		})
		position++
	}

	items = append(items, breadcrumbItem{
		Type:     "ListItem",
		Position: position,
		Name:     product.Name,
	})

	return items
}

func BuildProductStructuredData(
	product Product,
	canonicalURL string,
	imageURL string,
	priceAmount string,
	priceCurrency string,
	siteBaseURL string,
) json.RawMessage {
	type brand struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	}
	type offer struct {
		Type          string `json:"@type"`
		URL           string `json:"url"`
		PriceCurrency string `json:"priceCurrency"`
		Price         string `json:"price"`
		Availability  string `json:"availability"`
		ItemCondition string `json:"itemCondition"`
	}
	type breadcrumbList struct {
		Type     string           `json:"@type"`
		ItemList []breadcrumbItem `json:"itemListElement"`
	}
	type productSchema struct {
		Type        string   `json:"@type"`
		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		Image       []string `json:"image,omitempty"`
		SKU         string   `json:"sku"`
		URL         string   `json:"url"`
		Brand       brand    `json:"brand"`
		Offers      offer    `json:"offers"`
	}
	type graph struct {
		Context string `json:"@context"`
		Graph   []any  `json:"@graph"`
	}

	description := strings.TrimSpace(product.Description)
	images := []string{}
	if imageURL != "" {
		images = append(images, imageURL)
	}

	schema := graph{
		Context: "https://schema.org",
		Graph: []any{
			productSchema{
				Type:        "Product",
				Name:        fmt.Sprintf("%s %s", product.BrandName, product.Name),
				Description: description,
				Image:       images,
				SKU:         product.Serial,
				URL:         canonicalURL,
				Brand:       brand{Type: "Brand", Name: product.BrandName},
				Offers: offer{
					Type:          "Offer",
					URL:           canonicalURL,
					PriceCurrency: priceCurrency,
					Price:         priceAmount,
					Availability:  "https://schema.org/InStock",
					ItemCondition: "https://schema.org/NewCondition",
				},
			},
			breadcrumbList{
				Type:     "BreadcrumbList",
				ItemList: buildProductBreadcrumbItems(product, siteBaseURL),
			},
		},
	}

	data, err := json.Marshal(schema)
	if err != nil {
		return nil
	}
	return data
}

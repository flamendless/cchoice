package seo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	ProductRobots      = "index, follow, max-image-preview:large"
	ProductOGType        = "product"
	ProductTwitterCard   = "summary_large_image"
	DefaultOGImage       = "https://imagedelivery.net/YnES7emCTPeSEVA2N0dB_g/favicons-192x192/public"
	SitemapPlaceholder   = "# __SITEMAP__"
)

type Product struct {
	BrandName          string
	Name               string
	Serial             string
	Description        string
	ProductCategory    string
	ProductSubcategory string
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
	StructuredData string
}

type SitemapEntry struct {
	Loc        string
	LastMod    time.Time
	ChangeFreq string
	Priority   string
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

func GenerateProductMeta(
	product Product,
	canonicalURL string,
	homeURL string,
	imageURL string,
	priceAmount string,
	priceCurrency string,
) ProductMeta {
	title := fmt.Sprintf(
		"%s %s (%s) - Price, Specs, Buy Online | C-Choice",
		product.BrandName,
		product.Name,
		product.Serial,
	)

	description := strings.TrimSpace(product.Description)
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

func BuildProductStructuredData(
	product Product,
	canonicalURL string,
	imageURL string,
	priceAmount string,
	priceCurrency string,
	siteBaseURL string,
) string {
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

	description := strings.TrimSpace(product.Description)
	images := []string{}
	if imageURL != "" {
		images = append(images, imageURL)
	}

	items := []breadcrumbItem{
		{Type: "ListItem", Position: 1, Name: "Home", Item: strings.TrimSuffix(siteBaseURL, "/") + "/"},
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

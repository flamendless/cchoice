package seo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "ampersand", input: "a&b", want: "a&amp;b"},
		{name: "less than", input: "a<b", want: "a&lt;b"},
		{name: "greater than", input: "a>b", want: "a&gt;b"},
		{name: "quotes", input: `a"b'c`, want: "a&quot;b&apos;c"},
		{name: "url with query", input: "https://example.com?a=1&b=2", want: "https://example.com?a=1&amp;b=2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, EscapeXML(tt.input))
		})
	}
}

func TestInjectSitemapLine(t *testing.T) {
	content := "User-agent: *\nAllow: /\n\n# __SITEMAP__\n"
	got := InjectSitemapLine(content, "https://cchoice.shop/sitemap.xml")
	assert.Equal(t, "User-agent: *\nAllow: /\n\nSitemap: https://cchoice.shop/sitemap.xml\n", got)
}

func TestProductCanonicalURL(t *testing.T) {
	assert.Equal(
		t,
		"https://cchoice.shop/product/bosch-drill-123",
		ProductCanonicalURL("https://cchoice.shop", "bosch-drill-123"),
	)
}

func TestWriteSitemapURL(t *testing.T) {
	var buf bytes.Buffer

	WriteSitemapURL(&buf, SitemapEntry{
		Loc:        "https://cchoice.shop/product/test?a=1&b=2",
		LastMod:    time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC),
		ChangeFreq: "weekly",
		Priority:   "0.8",
	})

	got := buf.String()
	assert.Contains(t, got, "<loc>https://cchoice.shop/product/test?a=1&amp;b=2</loc>")
	assert.Contains(t, got, "<lastmod>2026-07-05</lastmod>")
	assert.Contains(t, got, "<changefreq>weekly</changefreq>")
	assert.Contains(t, got, "<priority>0.8</priority>")
}

func TestBuildSitemapXML(t *testing.T) {
	products := []SitemapEntry{
		{
			Loc:        "https://cchoice.shop/product/bosch-drill",
			LastMod:    time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			ChangeFreq: "weekly",
			Priority:   "0.8",
		},
	}

	got := BuildSitemapXML("https://cchoice.shop", products)

	assert.True(t, strings.HasPrefix(got, `<?xml version="1.0" encoding="UTF-8"?>`))
	assert.Contains(t, got, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	assert.Contains(t, got, "<loc>https://cchoice.shop</loc>")
	assert.Contains(t, got, "<loc>https://cchoice.shop/product/bosch-drill</loc>")
	assert.Contains(t, got, "<lastmod>2026-07-01</lastmod>")
	assert.Contains(t, got, `</urlset>`)
}

func gma55Product() Product {
	return Product{
		BrandName:          "Bosch",
		Name:               "GMA 55",
		Serial:             "BOSCH-GMA-163-0",
		ProductCategory:    "Table",
		ProductSubcategory: "Saw",
	}
}

const gma55Title = "Bosch GMA 55 Table Saw Power Tools | Price, Specs, Buy Online | C-Choice"

func TestGenerateProductMeta(t *testing.T) {
	product := gma55Product()
	product.Description = "Professional table saw for heavy-duty work."

	meta := GenerateProductMeta(
		product,
		"https://cchoice.shop/product/bosch-gma-55",
		"https://cchoice.shop",
		"https://cdn.example.com/product.webp",
		"12999.00",
		"PHP",
	)

	assert.Equal(t, gma55Title, meta.Title)
	assert.Equal(t, "Professional table saw for heavy-duty work.", meta.Description)
	assert.Equal(t, "https://cchoice.shop/product/bosch-gma-55", meta.CanonicalURL)
	assert.Equal(t, ProductOGType, meta.OGType)
	assert.Equal(t, ProductRobots, meta.Robots)
	assert.Equal(t, ProductTwitterCard, meta.TwitterCard)
	assert.Contains(t, meta.Keywords, "Bosch")
	assert.Contains(t, meta.Keywords, "Table")
	assert.Contains(t, meta.Keywords, "Saw")
	assert.Contains(t, meta.Keywords, "power tools")
	assert.Contains(t, meta.Keywords, "bosch philippines")
	assert.NotEmpty(t, meta.StructuredData)
}

func TestBuildProductTitle(t *testing.T) {
	title := buildProductTitle(gma55Product())
	assert.Equal(t, gma55Title, title)

	saleProduct := gma55Product()
	saleProduct.OnSale = true
	saleTitle := buildProductTitle(saleProduct)
	assert.Equal(t, "SALE! "+gma55Title, saleTitle)
}

func TestGenerateProductMeta_TruncatesLongDescription(t *testing.T) {
	longDescription := strings.Repeat("a", 200)
	product := gma55Product()
	product.Description = longDescription

	meta := GenerateProductMeta(
		product,
		"https://cchoice.shop/product/slug",
		"https://cchoice.shop",
		"",
		"100.00",
		"PHP",
	)

	assert.Len(t, meta.Description, 155)
	assert.True(t, strings.HasSuffix(meta.Description, "..."))
}

func TestGenerateProductMeta_FallbackDescription(t *testing.T) {
	product := gma55Product()

	meta := GenerateProductMeta(
		product,
		"https://cchoice.shop/product/bosch-gma-55",
		"https://cchoice.shop",
		"",
		"100.00",
		"PHP",
	)

	assert.Contains(t, meta.Description, "Shop Bosch GMA 55 (BOSCH-GMA-163-0)")
	assert.Contains(t, meta.Description, "C-Choice")
	assert.Contains(t, meta.Description, "Philippines")
	assert.Equal(t, DefaultOGImage, meta.OGImage)
}

func TestBuildProductStructuredData(t *testing.T) {
	product := gma55Product()
	product.Description = "Heavy-duty table saw"

	raw := BuildProductStructuredData(
		product,
		"https://cchoice.shop/product/bosch-gma-55",
		"https://cdn.example.com/product.webp",
		"12999.00",
		"PHP",
		"https://cchoice.shop",
	)

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &payload))

	assert.Equal(t, "https://schema.org", payload["@context"])
	assert.Equal(t, "Product", payload["@type"])
	assert.Equal(t, "Bosch GMA 55", payload["name"])
	assert.Equal(t, "BOSCH-GMA-163-0", payload["sku"])

	offers, ok := payload["offers"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "12999.00", offers["price"])
	assert.Equal(t, "PHP", offers["priceCurrency"])
	assert.Equal(t, "https://cchoice.shop/product/bosch-gma-55", offers["url"])

	breadcrumb, ok := payload["breadcrumb"].(map[string]any)
	require.True(t, ok)
	items, ok := breadcrumb["itemListElement"].([]any)
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(items), 4)

	tableItem, ok := items[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Table", tableItem["name"])

	sawItem, ok := items[2].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Saw", sawItem["name"])
}

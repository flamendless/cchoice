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

func TestGenerateProductMeta(t *testing.T) {
	product := Product{
		BrandName:          "Bosch",
		Name:               "Hammer Drill",
		Serial:             "GBH2-28",
		Description:        "Professional hammer drill for heavy-duty work.",
		ProductCategory:    "Power Tools",
		ProductSubcategory: "Drills",
	}

	meta := GenerateProductMeta(
		product,
		"https://cchoice.shop/product/bosch-hammer-drill-gbh2-28",
		"https://cchoice.shop",
		"https://cdn.example.com/product.webp",
		"12999.00",
		"PHP",
	)

	assert.Equal(t, "Bosch Hammer Drill (GBH2-28) - Price, Specs, Buy Online | C-Choice", meta.Title)
	assert.Equal(t, "Professional hammer drill for heavy-duty work.", meta.Description)
	assert.Equal(t, "https://cchoice.shop/product/bosch-hammer-drill-gbh2-28", meta.CanonicalURL)
	assert.Equal(t, ProductOGType, meta.OGType)
	assert.Equal(t, ProductRobots, meta.Robots)
	assert.Equal(t, ProductTwitterCard, meta.TwitterCard)
	assert.Contains(t, meta.Keywords, "Bosch")
	assert.Contains(t, meta.Keywords, "Power Tools")
	assert.NotEmpty(t, meta.StructuredData)
}

func TestGenerateProductMeta_TruncatesLongDescription(t *testing.T) {
	longDescription := strings.Repeat("a", 200)
	product := Product{
		BrandName:   "Bosch",
		Name:        "Hammer Drill",
		Serial:      "GBH2-28",
		Description: longDescription,
	}

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
	product := Product{
		BrandName: "Bosch",
		Name:      "Hammer Drill",
		Serial:    "GBH2-28",
	}

	meta := GenerateProductMeta(
		product,
		"https://cchoice.shop/product/slug",
		"https://cchoice.shop",
		"",
		"100.00",
		"PHP",
	)

	assert.Contains(t, meta.Description, "Shop Bosch Hammer Drill (GBH2-28)")
	assert.Equal(t, DefaultOGImage, meta.OGImage)
}

func TestBuildProductStructuredData(t *testing.T) {
	product := Product{
		BrandName:          "Bosch",
		Name:               "Hammer Drill",
		Serial:             "GBH2-28",
		Description:        "Heavy-duty drill",
		ProductCategory:    "Power Tools",
		ProductSubcategory: "Drills",
	}

	raw := BuildProductStructuredData(
		product,
		"https://cchoice.shop/product/bosch-hammer-drill",
		"https://cdn.example.com/product.webp",
		"12999.00",
		"PHP",
		"https://cchoice.shop",
	)

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &payload))

	assert.Equal(t, "https://schema.org", payload["@context"])
	assert.Equal(t, "Product", payload["@type"])
	assert.Equal(t, "Bosch Hammer Drill", payload["name"])
	assert.Equal(t, "GBH2-28", payload["sku"])

	offers, ok := payload["offers"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "12999.00", offers["price"])
	assert.Equal(t, "PHP", offers["priceCurrency"])
	assert.Equal(t, "https://cchoice.shop/product/bosch-hammer-drill", offers["url"])

	breadcrumb, ok := payload["breadcrumb"].(map[string]any)
	require.True(t, ok)
	items, ok := breadcrumb["itemListElement"].([]any)
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(items), 4)
}

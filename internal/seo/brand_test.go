package seo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrandsListingPageURL(t *testing.T) {
	assert.Equal(t, "https://cchoice.shop/brands/", BrandsListingPageURL("https://cchoice.shop"))
}

func TestBrandPageURL(t *testing.T) {
	assert.Equal(t, "https://cchoice.shop/brands/bosch", BrandPageURL("https://cchoice.shop", "bosch"))
}

func TestBuildBrandsListingStructuredData(t *testing.T) {
	canonicalURL := "https://cchoice.shop/brands/"
	title := "Brands | C-Choice Construction Supply"
	description := "Browse all power tool and construction supply brands."

	raw := BuildBrandsListingStructuredData(canonicalURL, title, description)
	require.NotEmpty(t, raw)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(raw, &payload))

	graph, ok := payload["@graph"].([]any)
	require.True(t, ok)
	require.Len(t, graph, 2)

	collectionPage, ok := graph[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "CollectionPage", collectionPage["@type"])
	assert.Equal(t, title, collectionPage["name"])
	assert.Equal(t, canonicalURL, collectionPage["url"])

	breadcrumb, ok := graph[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "BreadcrumbList", breadcrumb["@type"])

	items, ok := breadcrumb["itemListElement"].([]any)
	require.True(t, ok)
	require.Len(t, items, 2)

	homeItem, ok := items[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "https://cchoice.shop/", homeItem["item"])

	brandsItem, ok := items[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Brands", brandsItem["name"])
	assert.Equal(t, canonicalURL, brandsItem["item"])
}

func TestBuildBrandStructuredData(t *testing.T) {
	canonicalURL := "https://cchoice.shop/brands/bosch"
	title := "Bosch | C-Choice Construction Supply"
	description := "Shop Bosch power tools and construction supplies."

	raw := BuildBrandStructuredData("bosch", "Bosch", canonicalURL, title, description)
	require.NotEmpty(t, raw)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(raw, &payload))

	graph, ok := payload["@graph"].([]any)
	require.True(t, ok)
	require.Len(t, graph, 2)

	breadcrumb, ok := graph[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "BreadcrumbList", breadcrumb["@type"])

	items, ok := breadcrumb["itemListElement"].([]any)
	require.True(t, ok)
	require.Len(t, items, 3)

	brandsItem, ok := items[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Brands", brandsItem["name"])
	assert.Equal(t, "https://cchoice.shop/brands/", brandsItem["item"])

	brandItem, ok := items[2].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Bosch", brandItem["name"])
	assert.Equal(t, canonicalURL, brandItem["item"])
}

func TestBuildProductBreadcrumbItems_WithBrandSlug(t *testing.T) {
	items := buildProductBreadcrumbItems(
		Product{
			BrandName:          "Bosch",
			BrandSlug:          "bosch",
			Name:               "GMA 55",
			ProductCategory:    "table",
			ProductSubcategory: "saw",
		},
		"https://cchoice.shop",
	)

	require.Len(t, items, 5)
	assert.Equal(t, "Bosch", items[1].Name)
	assert.Equal(t, "https://cchoice.shop/brands/bosch", items[1].Item)
	assert.Equal(t, "Table", items[2].Name)
	assert.Equal(t, "Saw", items[3].Name)
	assert.Equal(t, "GMA 55", items[4].Name)
}

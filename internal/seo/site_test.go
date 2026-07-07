package seo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSiteStructuredData(t *testing.T) {
	raw := BuildSiteStructuredData("https://cchoice.shop", DefaultOGImage)
	require.NotEmpty(t, raw)

	var payload struct {
		Context string           `json:"@context"`
		Graph   []map[string]any `json:"@graph"`
	}
	require.NoError(t, json.Unmarshal(raw, &payload))

	assert.Equal(t, "https://schema.org", payload.Context)
	require.Len(t, payload.Graph, 2)

	org := payload.Graph[0]
	assert.Equal(t, "Organization", org["@type"])
	assert.Equal(t, "C-Choice Construction Supply", org["name"])
	assert.Equal(t, "Philippines", org["areaServed"])

	alternateNames, ok := org["alternateName"].([]any)
	require.True(t, ok)
	assert.Contains(t, alternateNames, "cchoice")

	website := payload.Graph[1]
	assert.Equal(t, "WebSite", website["@type"])
	assert.Equal(t, "https://cchoice.shop/", website["url"])
}

func TestBaseProductKeywords(t *testing.T) {
	keywords := BaseProductKeywords()
	assert.Contains(t, keywords, "power tools")
	assert.Contains(t, keywords, "construction supplies")
	assert.Contains(t, keywords, "philippines")
	assert.Contains(t, keywords, "cchoice")
}

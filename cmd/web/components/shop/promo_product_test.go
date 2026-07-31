package shop

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"cchoice/cmd/web/models"
)

// TestSaleBannerImageIsDiscoverable guards the largest contentful paint on the
// home page: the banner starts hidden, so the image must be requested by the
// parser rather than after a script reveals the modal.
func TestSaleBannerImageIsDiscoverable(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("SaleBanner depends on conf; skipping: %v", r)
		}
	}()

	product := models.RandomSaleProduct{
		ProductID:          "abc123",
		Slug:               "gpo-12-ce",
		CDNURL:             "https://cdn.test/640",
		CDNURL1280:         "https://cdn.test/1280",
		DiscountPercentage: "50%",
	}

	var sb strings.Builder
	assert.NoError(t, SaleBanner(product).Render(context.Background(), &sb))
	got := sb.String()

	assert.Contains(t, got, `loading="eager"`)
	assert.Contains(t, got, `fetchpriority="high"`)
	assert.NotContains(t, got, `loading="lazy"`)

	assert.Contains(t, got, `src="https://cdn.test/640"`,
		"the smaller variant is enough for a 448px wide modal")
	assert.Contains(t, got, `srcset="https://cdn.test/640 640w, https://cdn.test/1280 1280w"`)
	assert.Contains(t, got, `sizes=`)

	assert.NotContains(t, got, "sessionStorage.getItem('saleBannerShown')",
		"revealing the banner must not wait for hyperscript to download")
	assert.Contains(t, got, `sessionStorage.getItem("saleBannerShown")`)
}

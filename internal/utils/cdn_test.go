package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"cchoice/internal/conf"
)

func withFlexibleVariants(t *testing.T, enabled bool) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("CDN helpers depend on conf; skipping: %v", r)
		}
	}()

	cfg := conf.Conf()
	previous := cfg.CloudflareImages.FlexibleVariants
	cfg.CloudflareImages.FlexibleVariants = enabled
	t.Cleanup(func() { cfg.CloudflareImages.FlexibleVariants = previous })
}

func TestCDNImageWidth(t *testing.T) {
	const delivery = "https://imagedelivery.net/acct/brand_logos-BOSCH/public"

	t.Run("disabled leaves the url untouched", func(t *testing.T) {
		withFlexibleVariants(t, false)
		assert.Equal(t, delivery, CDNImageWidth(delivery, 512))
	})

	t.Run("enabled", func(t *testing.T) {
		withFlexibleVariants(t, true)

		tests := []struct {
			name     string
			rawURL   string
			width    int
			expected string
		}{
			{
				name:     "replaces the variant with a resize transform",
				rawURL:   delivery,
				width:    512,
				expected: "https://imagedelivery.net/acct/brand_logos-BOSCH/w=512,fit=scale-down,format=auto",
			},
			{
				name:     "non-positive width",
				rawURL:   delivery,
				width:    0,
				expected: delivery,
			},
			{
				name:     "already transformed",
				rawURL:   "https://imagedelivery.net/acct/brand_logos-BOSCH/w=100",
				width:    512,
				expected: "https://imagedelivery.net/acct/brand_logos-BOSCH/w=100",
			},
			{
				name:     "not a delivery url",
				rawURL:   "/products/image?path=static/images/a.webp",
				width:    512,
				expected: "/products/image?path=static/images/a.webp",
			},
			{
				name:     "empty",
				rawURL:   "",
				width:    512,
				expected: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, CDNImageWidth(tt.rawURL, tt.width))
			})
		}
	})
}

func TestCDNImageSrcSet(t *testing.T) {
	const delivery = "https://imagedelivery.net/acct/store/public"

	t.Run("disabled yields no srcset", func(t *testing.T) {
		withFlexibleVariants(t, false)
		assert.Empty(t, CDNImageSrcSet(delivery, 512, 1024))
	})

	t.Run("enabled", func(t *testing.T) {
		withFlexibleVariants(t, true)

		assert.Equal(
			t,
			"https://imagedelivery.net/acct/store/w=512,fit=scale-down,format=auto 512w, "+
				"https://imagedelivery.net/acct/store/w=1024,fit=scale-down,format=auto 1024w",
			CDNImageSrcSet(delivery, 512, 1024),
		)
		assert.Empty(t, CDNImageSrcSet(delivery))
		assert.Empty(t, CDNImageSrcSet("https://example.test/a.png", 512))
	})
}

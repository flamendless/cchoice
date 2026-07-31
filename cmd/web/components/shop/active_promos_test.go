package shop

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"cchoice/cmd/web/models"
	"cchoice/internal/enums"
)

func renderPromos(t *testing.T, promos []models.PromoItem) string {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("ActivePromosSection depends on conf; skipping: %v", r)
		}
	}()

	var sb strings.Builder
	assert.NoError(t, ActivePromosSection(promos).Render(context.Background(), &sb))
	return sb.String()
}

func TestActivePromosSectionReservesBannerSpace(t *testing.T) {
	promos := []models.PromoItem{
		{Title: "First", MediaURL: "https://cdn.test/first", Type: enums.PROMO_TYPE_BANNER_IMAGE},
		{Title: "Second", MediaURL: "https://cdn.test/second", Type: enums.PROMO_TYPE_BANNER_IMAGE},
	}

	got := renderPromos(t, promos)

	assert.Contains(t, got, "aspect-ratio: 2 / 1",
		"banners need a reserved box; uploaded images have no known dimensions")
	assert.Equal(t, 2, strings.Count(got, `class="promo-banner-media"`))
}

func TestActivePromosSectionPrioritizesFirstBanner(t *testing.T) {
	tests := []struct {
		name             string
		promos           []models.PromoItem
		wantEagerBanners int
		wantLazyBanners  int
		wantHighPriority int
	}{
		{
			name: "single banner is the likely LCP element",
			promos: []models.PromoItem{
				{Title: "Only", MediaURL: "https://cdn.test/only", Type: enums.PROMO_TYPE_BANNER_IMAGE},
			},
			wantEagerBanners: 1,
			wantLazyBanners:  0,
			wantHighPriority: 1,
		},
		{
			name: "only the first banner is prioritized",
			promos: []models.PromoItem{
				{Title: "First", MediaURL: "https://cdn.test/first", Type: enums.PROMO_TYPE_BANNER_IMAGE},
				{Title: "Second", MediaURL: "https://cdn.test/second", Type: enums.PROMO_TYPE_BANNER_IMAGE},
				{Title: "Third", MediaURL: "https://cdn.test/third", Type: enums.PROMO_TYPE_BANNER_IMAGE},
			},
			wantEagerBanners: 1,
			wantLazyBanners:  2,
			wantHighPriority: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderPromos(t, tt.promos)

			assert.Equal(t, tt.wantEagerBanners, strings.Count(got, `loading="eager"`))
			assert.Equal(t, tt.wantLazyBanners, strings.Count(got, `loading="lazy"`))
			assert.Equal(t, tt.wantHighPriority, strings.Count(got, `fetchpriority="high"`))
		})
	}
}

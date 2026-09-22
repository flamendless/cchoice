package product

import (
	"testing"

	"cchoice/cmd/web/models"

	"github.com/stretchr/testify/assert"
)

func TestHeroImageURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data models.ProductPageData
		want string
	}{
		{
			name: "prefers og image",
			data: models.ProductPageData{
				Meta:       models.ProductsMeta{OGImage: "https://cdn.example/product_images-bosch-original.jpg"},
				CDNURL:     "https://cdn.example/product_images-bosch-webp-1280x1280.jpg",
				CDNURL1280: "https://cdn.example/thumb.jpg",
			},
			want: "https://cdn.example/product_images-bosch-original.jpg",
		},
		{
			name: "falls back to stored cdn url",
			data: models.ProductPageData{
				CDNURL:     "https://cdn.example/product_images-bosch-webp-1280x1280.jpg",
				CDNURL1280: "https://cdn.example/thumb.jpg",
			},
			want: "https://cdn.example/product_images-bosch-webp-1280x1280.jpg",
		},
		{
			name: "falls back to cdn url thumbnail",
			data: models.ProductPageData{
				CDNURL1280: "https://cdn.example/thumb.jpg",
			},
			want: "https://cdn.example/thumb.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, heroImageURL(tt.data))
		})
	}
}

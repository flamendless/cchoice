package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveSEOImageURL(t *testing.T) {
	t.Parallel()

	cdn := func(path string) string {
		if path == "" {
			return ""
		}
		return "https://cdn.example.com" + path
	}
	svc := &ProductService{getCDNURL: cdn}

	tests := []struct {
		name            string
		cdnURL          string
		cdnURLThumbnail string
		imagePath       string
		thumbnailPath   string
		want            string
	}{
		{
			name:      "prefers original image path",
			imagePath: "/products/original.webp",
			cdnURL:    "https://cdn.example.com/products/1280.webp",
			want:      "https://cdn.example.com/products/original.webp",
		},
		{
			name:   "falls back to cdn url",
			cdnURL: "https://cdn.example.com/products/full.webp",
			want:   "https://cdn.example.com/products/full.webp",
		},
		{
			name:            "falls back to thumbnail cdn url",
			cdnURLThumbnail: "https://cdn.example.com/products/thumb.webp",
			want:            "https://cdn.example.com/products/thumb.webp",
		},
		{
			name:          "falls back to 1280 thumbnail path",
			thumbnailPath: "/products/500x500/item.webp",
			want:          "https://cdn.example.com/products/1280x1280/item.webp",
		},
		{
			name: "returns empty when no image data",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.resolveSEOImageURL(tt.cdnURL, tt.cdnURLThumbnail, tt.imagePath, tt.thumbnailPath)
			assert.Equal(t, tt.want, got)
		})
	}
}

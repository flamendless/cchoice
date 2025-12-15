package images

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var tblImageFormat = map[ImageFormat]string{
	IMAGE_FORMAT_UNDEFINED: "UNDEFINED",
	IMAGE_FORMAT_PNG:       "PNG",
	IMAGE_FORMAT_WEBP:      "WEBP",
	IMAGE_FORMAT_JPEG:      "JPEG",
	IMAGE_FORMAT_GIF:       "GIF",
	IMAGE_FORMAT_SVG:       "SVG",
	IMAGE_FORMAT_BMP:       "BMP",
	IMAGE_FORMAT_ICO:       "ICO",
}

var tblImageFormatExt = map[ImageFormat][]string{
	IMAGE_FORMAT_PNG:  {"png", ".png"},
	IMAGE_FORMAT_WEBP: {"webp", ".webp"},
	IMAGE_FORMAT_JPEG: {"jpg", ".jpg", "jpeg", ".jpeg"},
	IMAGE_FORMAT_GIF:  {"gif", ".gif"},
	IMAGE_FORMAT_SVG:  {"svg", ".svg"},
	IMAGE_FORMAT_BMP:  {"bmp", ".bmp"},
	IMAGE_FORMAT_ICO:  {"ico", ".ico"},
}

func TestImageFormatToString(t *testing.T) {
	for imageFormat, val := range tblImageFormat {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, val, imageFormat.String())
		})
	}
}

func TestParseImageFormatEnum(t *testing.T) {
	for imageFormat, val := range tblImageFormat {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, imageFormat, ParseImageFormatToEnum(val))
		})
	}
}

func TestParseImageFormatExtEnum(t *testing.T) {
	for imageFormatExt, val := range tblImageFormatExt {
		t.Run(imageFormatExt.String(), func(t *testing.T) {
			for _, ext := range val {
				require.Equal(t, imageFormatExt, ParseImageFormatExtToEnum(ext))
			}
		})
	}
}

func BenchmarkImageFormatToString(b *testing.B) {
	for imageFormat := range tblImageFormat {
		b.Run(imageFormat.String(), func(b *testing.B) {
			for b.Loop() {
				_ = imageFormat.String()
			}
		})
	}
}

func BenchmarkParseImageFormatEnum(b *testing.B) {
	for _, val := range tblImageFormat {
		b.Run(val, func(b *testing.B) {
			for b.Loop() {
				_ = ParseImageFormatToEnum(val)
			}
		})
	}
}

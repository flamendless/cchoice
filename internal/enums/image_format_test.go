package enums

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var tblImageFormat = map[ImageFormat]string{
	IMAGE_FORMAT_UNDEFINED: "UNDEFINED",
	IMAGE_FORMAT_PNG:       "PNG",
	IMAGE_FORMAT_WEBP:      "WEBP",
}

var tblImageFormatExt = map[ImageFormat][]string{
	IMAGE_FORMAT_PNG:  {"png", ".png"},
	IMAGE_FORMAT_WEBP: {"webp", ".webp"},
}

func TestImageFormatToString(t *testing.T) {
	for imageFormat, val := range tblImageFormat {
		t.Run(val, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, val, imageFormat.String())
		})
	}
}

func TestParseImageFormatEnum(t *testing.T) {
	for imageFormat, val := range tblImageFormat {
		t.Run(val, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, imageFormat, ParseImageFormatToEnum(val))
		})
	}
}

func TestParseImageFormatExtEnum(t *testing.T) {
	for imageFormatExt, val := range tblImageFormatExt {
		t.Run(imageFormatExt.String(), func(t *testing.T) {
			t.Parallel()
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

package images

import "strings"

//go:generate go tool stringer -type=ImageFormat -trimprefix=IMAGE_FORMAT_

type ImageFormat int

const (
	IMAGE_FORMAT_UNDEFINED ImageFormat = iota
	IMAGE_FORMAT_PNG
	IMAGE_FORMAT_WEBP
	IMAGE_FORMAT_JPEG
	IMAGE_FORMAT_GIF
	IMAGE_FORMAT_SVG
	IMAGE_FORMAT_BMP
	IMAGE_FORMAT_ICO
)

var AllImageFormats = []ImageFormat{
	IMAGE_FORMAT_PNG,
	IMAGE_FORMAT_WEBP,
	IMAGE_FORMAT_JPEG,
	IMAGE_FORMAT_GIF,
	IMAGE_FORMAT_SVG,
	IMAGE_FORMAT_BMP,
	IMAGE_FORMAT_ICO,
}

func ParseImageFormatToEnum(format string) ImageFormat {
	switch strings.ToUpper(format) {
	case IMAGE_FORMAT_PNG.String():
		return IMAGE_FORMAT_PNG
	case IMAGE_FORMAT_WEBP.String():
		return IMAGE_FORMAT_WEBP
	case IMAGE_FORMAT_JPEG.String(), "JPG":
		return IMAGE_FORMAT_JPEG
	case IMAGE_FORMAT_GIF.String():
		return IMAGE_FORMAT_GIF
	case IMAGE_FORMAT_SVG.String():
		return IMAGE_FORMAT_SVG
	case IMAGE_FORMAT_BMP.String():
		return IMAGE_FORMAT_BMP
	case IMAGE_FORMAT_ICO.String():
		return IMAGE_FORMAT_ICO
	default:
		return IMAGE_FORMAT_UNDEFINED
	}
}

func ParseImageFormatExtToEnum(format string) ImageFormat {
	switch strings.ToLower(format) {
	case ".png", "png":
		return IMAGE_FORMAT_PNG
	case ".webp", "webp":
		return IMAGE_FORMAT_WEBP
	case ".jpg", "jpg", ".jpeg", "jpeg":
		return IMAGE_FORMAT_JPEG
	case ".gif", "gif":
		return IMAGE_FORMAT_GIF
	case ".svg", "svg":
		return IMAGE_FORMAT_SVG
	case ".bmp", "bmp":
		return IMAGE_FORMAT_BMP
	case ".ico", "ico":
		return IMAGE_FORMAT_ICO
	default:
		return IMAGE_FORMAT_UNDEFINED
	}
}

func (f ImageFormat) Extension() string {
	switch f {
	case IMAGE_FORMAT_PNG:
		return ".png"
	case IMAGE_FORMAT_WEBP:
		return ".webp"
	case IMAGE_FORMAT_JPEG:
		return ".jpeg"
	case IMAGE_FORMAT_GIF:
		return ".gif"
	case IMAGE_FORMAT_SVG:
		return ".svg"
	case IMAGE_FORMAT_BMP:
		return ".bmp"
	case IMAGE_FORMAT_ICO:
		return ".ico"
	default:
		return ""
	}
}

func (f ImageFormat) MIMEType() string {
	switch f {
	case IMAGE_FORMAT_PNG:
		return "image/png"
	case IMAGE_FORMAT_WEBP:
		return "image/webp"
	case IMAGE_FORMAT_JPEG:
		return "image/jpeg"
	case IMAGE_FORMAT_GIF:
		return "image/gif"
	case IMAGE_FORMAT_SVG:
		return "image/svg+xml"
	case IMAGE_FORMAT_BMP:
		return "image/bmp"
	case IMAGE_FORMAT_ICO:
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}

func (f ImageFormat) DataURIPrefix() string {
	switch f {
	case IMAGE_FORMAT_PNG:
		return "data:image/png;base64,"
	case IMAGE_FORMAT_WEBP:
		return "data:image/webp;base64,"
	case IMAGE_FORMAT_JPEG:
		return "data:image/jpeg;base64,"
	case IMAGE_FORMAT_GIF:
		return "data:image/gif;base64,"
	case IMAGE_FORMAT_SVG:
		return "data:image/svg+xml;base64,"
	case IMAGE_FORMAT_BMP:
		return "data:image/bmp;base64,"
	case IMAGE_FORMAT_ICO:
		return "data:image/x-icon;base64,"
	default:
		return "data:application/octet-stream;base64,"
	}
}

func IsValidImageExtension(ext string) bool {
	return ParseImageFormatExtToEnum(ext) != IMAGE_FORMAT_UNDEFINED
}

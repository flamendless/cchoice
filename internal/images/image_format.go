package images

//go:generate go tool stringer -type=ImageFormat -trimprefix=IMAGE_FORMAT_

type ImageFormat int

const (
	IMAGE_FORMAT_UNDEFINED ImageFormat = iota
	IMAGE_FORMAT_PNG
	IMAGE_FORMAT_WEBP
)

func ParseImageFormatToEnum(format string) ImageFormat {
	switch format {
	case IMAGE_FORMAT_PNG.String():
		return IMAGE_FORMAT_PNG
	case IMAGE_FORMAT_WEBP.String():
		return IMAGE_FORMAT_WEBP
	default:
		return IMAGE_FORMAT_UNDEFINED
	}
}

func ParseImageFormatExtToEnum(format string) ImageFormat {
	switch format {
	case ".png", "png":
		return IMAGE_FORMAT_PNG
	case ".webp", "webp":
		return IMAGE_FORMAT_WEBP
	default:
		return IMAGE_FORMAT_UNDEFINED
	}
}

func (f ImageFormat) MIMEType() string {
	switch f {
	case IMAGE_FORMAT_PNG:
		return "image/png"
	case IMAGE_FORMAT_WEBP:
		return "image/webp"
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
	default:
		return "data:application/octet-stream;base64,"
	}
}

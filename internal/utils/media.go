package utils

import (
	"cchoice/internal/enums"
)

func GetContentType(ext string) string {
	imgFormat := enums.ParseImageFormatExtToEnum(ext)
	if imgFormat != enums.IMAGE_FORMAT_UNDEFINED {
		return imgFormat.MIMEType()
	}

	switch ext {
	case ".md":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	case ".html":
		return "text/html"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

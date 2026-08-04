package utils

import (
	"fmt"
	"net/http"

	"cchoice/internal/constants"
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

func IsYouTubeURL(url string) bool {
	for _, pattern := range constants.YoutubePatterns {
		if pattern.MatchString(url) {
			return true
		}
	}
	return false
}

func IsValidProductImageContentType(contentType string) bool {
	switch contentType {
	case enums.IMAGE_FORMAT_JPEG.MIMEType(), enums.IMAGE_FORMAT_PNG.MIMEType(), enums.IMAGE_FORMAT_WEBP.MIMEType():
		return true
	default:
		return false
	}
}

func DetectProductImageContentType(data []byte) (string, error) {
	contentType := http.DetectContentType(data)
	if !IsValidProductImageContentType(contentType) {
		return "", fmt.Errorf("invalid image content type: %s", contentType)
	}
	return contentType, nil
}

func ConvertYouTubeToEmbed(url string) string {
	for _, pattern := range constants.YoutubePatterns {
		matches := pattern.FindStringSubmatch(url)
		if len(matches) >= 2 {
			videoID := matches[1]
			return "https://www.youtube.com/embed/" + videoID
		}
	}
	return url
}

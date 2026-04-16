package utils

import (
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

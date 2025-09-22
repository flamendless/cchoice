package utils

import (
	"slices"
	"strings"
)

func IsValidImageExtension(ext string) bool {
	validExts := []string{".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp"}
	ext = strings.ToLower(ext)
	return slices.Contains(validExts, ext)
}

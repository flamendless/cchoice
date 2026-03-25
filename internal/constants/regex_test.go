package constants

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexVariables(t *testing.T) {
	tests := []struct {
		name     string
		regex    *regexp.Regexp
		isValid  bool
		testCase string
	}{
		{"ReSize valid", ReSize, true, "/500x500/"},
		{"ReSize invalid", ReSize, false, "500x500"},
		{"ReSize invalid", ReSize, false, "notasize"},
		{"ReMultipleSpaces valid", ReMultipleSpaces, true, "many   spaces"},
		{"ReOrderReference valid", ReOrderReference, true, "CCO-abc1234A1B2C3"},
		{"ReOrderReference invalid", ReOrderReference, false, "CCO-abc123"},
		{"RePassword valid", RePassword, true, "Password123-_.?#@"},
		{"RePassword invalid", RePassword, false, "pass word"},
		{"ReEmail valid", ReEmail, true, "test@example.com"},
		{"ReEmail invalid", ReEmail, false, "invalid-email"},
		{"ReName valid", ReName, true, "John Doe"},
		{"ReName invalid", ReName, false, "John123"},
		{"ReMobileNumber valid", ReMobileNumber, true, "09123456789"},
		{"ReMobileNumber invalid", ReMobileNumber, false, "0912345678"},
		{"RePostalCode valid", RePostalCode, true, "1234"},
		{"RePostalCode invalid", RePostalCode, false, "12345"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := tt.regex.MatchString(tt.testCase)
			if tt.isValid {
				assert.True(t, matched, "expected %q to match", tt.testCase)
			} else {
				assert.False(t, matched, "expected %q not to match", tt.testCase)
			}
		})
	}
}

func TestToPath1280(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"standard size", "images/500x500/product.jpg", "images/1280x1280/product.jpg"},
		{"multiple sizes replaces first only", "100x100/200x200/image.png", "100x100/1280x1280/image.png"},
		{"no size", "images/product.jpg", "images/product.jpg"},
		{"empty string", "", ""},
		{"large dimensions", "1920x1080/image.jpg", "1920x1080/image.jpg"},
		{"underscore filename", "100x100_myimage.jpg", "100x100_myimage.jpg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToPath1280(tt.input))
		})
	}
}

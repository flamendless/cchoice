package utils

import "testing"

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no changes",
			input:    "GWS 700",
			expected: "GWS 700",
		},
		{
			name:     "unix newline",
			input:    "GSR 12V-15FC\nFlexiClick 5-in-1",
			expected: "GSR 12V-15FC FlexiClick 5-in-1",
		},
		{
			name:     "windows newline",
			input:    "GSR 12V-15FC\r\nFlexiClick 5-in-1",
			expected: "GSR 12V-15FC FlexiClick 5-in-1",
		},
		{
			name:     "old mac newline",
			input:    "GSR 12V-15FC\rFlexiClick 5-in-1",
			expected: "GSR 12V-15FC FlexiClick 5-in-1",
		},
		{
			name:     "multiple newlines",
			input:    "GSR 12V-15FC\n\nFlexiClick  5-in-1",
			expected: "GSR 12V-15FC FlexiClick 5-in-1",
		},
		{
			name:     "leading and trailing whitespace",
			input:    "  GSR 12V-15FC\nFlexiClick 5-in-1  ",
			expected: "GSR 12V-15FC FlexiClick 5-in-1",
		},
		{
			name:     "only whitespace and newlines",
			input:    "\n  \r\n  ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeName(tt.input)
			if result != tt.expected {
				t.Fatalf("SanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatAdminOrderAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "joins non-empty parts",
			parts:    []string{"123 Main St", "Manila", "Philippines"},
			expected: "123 Main St, Manila, Philippines",
		},
		{
			name:     "skips empty parts",
			parts:    []string{"123 Main St", "", "Manila", "", "Philippines"},
			expected: "123 Main St, Manila, Philippines",
		},
		{
			name:     "trims trailing commas",
			parts:    []string{"123 Main St", "", ""},
			expected: "123 Main St",
		},
		{
			name:     "empty address",
			parts:    []string{"", "", ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, formatAdminOrderAddress(tt.parts...))
		})
	}
}

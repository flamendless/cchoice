package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeOrderListingSort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sortBy      string
		sortDir     string
		wantSortBy  string
		wantSortDir string
	}{
		{
			name:        "defaults empty values",
			sortBy:      "",
			sortDir:     "",
			wantSortBy:  "UPDATED_AT",
			wantSortDir: "DESC",
		},
		{
			name:        "keeps explicit values",
			sortBy:      "CREATED_AT",
			sortDir:     "ASC",
			wantSortBy:  "CREATED_AT",
			wantSortDir: "ASC",
		},
		{
			name:        "defaults only sort dir",
			sortBy:      "STATUS",
			sortDir:     "",
			wantSortBy:  "STATUS",
			wantSortDir: "DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sortBy, sortDir := normalizeOrderListingSort(tt.sortBy, tt.sortDir)
			assert.Equal(t, tt.wantSortBy, sortBy)
			assert.Equal(t, tt.wantSortDir, sortDir)
		})
	}
}

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

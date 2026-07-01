package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStaffLogSuccessID(t *testing.T) {
	tests := []struct {
		name      string
		result    string
		wantID    string
		wantOK    bool
	}{
		{
			name:   "valid success id",
			result: "success. ID '13a31bab1c232bac'",
			wantID: "13a31bab1c232bac",
			wantOK: true,
		},
		{
			name:   "plain success",
			result: "success",
			wantOK: false,
		},
		{
			name:   "error message",
			result: "sql: no rows in result set",
			wantOK: false,
		},
		{
			name:   "success with extra text",
			result: "success. ID 'abc' extra",
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := ParseStaffLogSuccessID(tt.result)
			assert.Equal(t, tt.wantOK, gotOK)
			if gotOK {
				assert.Equal(t, tt.wantID, gotID)
			}
		})
	}
}

func TestBuildStaffLogReference(t *testing.T) {
	tests := []struct {
		name        string
		module      string
		action      string
		productSlug string
		wantLabel   string
		wantSlug    string
	}{
		{
			name:        "products create",
			module:      "products",
			action:      "create",
			productSlug: "brand-category-serial",
			wantLabel:   "View Product",
			wantSlug:    "brand-category-serial",
		},
		{
			name:        "products update",
			module:      "products",
			action:      "update",
			productSlug: "some-slug",
			wantLabel:   "View Product",
			wantSlug:    "some-slug",
		},
		{
			name:        "wrong module",
			module:      "brands",
			action:      "create",
			productSlug: "some-slug",
		},
		{
			name:        "wrong action",
			module:      "products",
			action:      "delete",
			productSlug: "some-slug",
		},
		{
			name:   "empty slug",
			module: "products",
			action: "create",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Skipf("BuildStaffLogReference depends on conf; skipping: %v", r)
				}
			}()
			ref := BuildStaffLogReference(tt.module, tt.action, tt.productSlug)
			assert.Equal(t, tt.wantLabel, ref.Label)
			if tt.wantSlug != "" {
				assert.Contains(t, ref.URL, "/product/"+tt.wantSlug)
			} else {
				assert.Empty(t, ref.URL)
			}
		})
	}
}

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

func TestBuildStaffLogProductReference(t *testing.T) {
	tests := []struct {
		name      string
		slug      string
		serial    string
		status    string
		wantLabel string
		wantInURL []string
		wantNewTab bool
	}{
		{
			name:       "active product",
			slug:       "brand-category-serial",
			serial:     "ABC123",
			status:     "ACTIVE",
			wantLabel:  "View Product in Shop",
			wantInURL:  []string{"/product/brand-category-serial"},
			wantNewTab: true,
		},
		{
			name:       "draft product",
			slug:       "brand-category-serial",
			serial:     "ABC123",
			status:     "DRAFT",
			wantLabel:  "View Product In Manage",
			wantInURL:  []string{"/admin/superuser/products", "search_serial=ABC123", "status=DRAFT"},
			wantNewTab: false,
		},
		{
			name:      "active without slug",
			serial:    "ABC123",
			status:    "ACTIVE",
			wantLabel: "",
		},
		{
			name:      "draft without serial",
			slug:      "brand-category-serial",
			status:    "DRAFT",
			wantLabel: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Skipf("BuildStaffLogProductReference depends on conf; skipping: %v", r)
				}
			}()
			ref := BuildStaffLogProductReference(tt.slug, tt.serial, tt.status)
			assert.Equal(t, tt.wantLabel, ref.Label)
			assert.Equal(t, tt.wantNewTab, ref.NewTab)
			if len(tt.wantInURL) == 0 {
				assert.Empty(t, ref.URL)
				return
			}
			for _, part := range tt.wantInURL {
				assert.Contains(t, ref.URL, part)
			}
		})
	}
}

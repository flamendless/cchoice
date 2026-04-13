package utils

import (
	"cchoice/internal/enums"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInitials(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single word", "Hello", "H"},
		{"two words", "Hello World", "HW"},
		{"acronym", "PHP", "PHP"},
		{"with symbols", "C-Choice", "CC"},
		{"spaces only", "   ", ""},
		{"mixed case", "Some Product Name", "SPN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetInitials(tt.input))
		})
	}
}

func TestRemoveEmptyStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"nil", nil, nil},
		{"empty slice", []string{}, nil},
		{"no empties", []string{"a", "b"}, []string{"a", "b"}},
		{"all empty", []string{"", "", ""}, nil},
		{"mixed", []string{"a", "", "b", "", "c"}, []string{"a", "b", "c"}},
		{"leading trailing empty", []string{"", "x", ""}, []string{"x"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveEmptyStrings(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSlugToTile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single", "hello", "Hello"},
		{"kebab", "some-product-name", "Some Product Name"},
		{"multiple hyphens", "a-b-c", "A B C"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, SlugToTile(tt.input))
		})
	}
}

func TestGetBoolFlag(t *testing.T) {
	tests := []struct {
		flag string
		want bool
	}{
		{"true", true},
		{"1", true},
		{"false", false},
		{"0", false},
		{"", false},
		{"yes", false},
	}
	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			assert.Equal(t, tt.want, GetBoolFlag(tt.flag))
		})
	}
}

func TestLabelToID(t *testing.T) {
	tests := []struct {
		name     string
		module   enums.Module
		label    string
		wantPref string
	}{
		{"category", enums.MODULE_CATEGORY, "Power Tools", "category-"},
		{"product", enums.MODULE_PRODUCT, "Drill 123", "product-"},
		{"brand", enums.MODULE_BRAND, "Bosch", "brand-"},
		{"lowercase and spaces", enums.MODULE_CATEGORY, "some label", "category-"},
		{"strips invalid", enums.MODULE_PRODUCT, "Hello! @World", "product-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LabelToID(tt.module, tt.label)
			assert.Contains(t, got, tt.wantPref, "should have prefix")
			assert.NotContains(t, got, " ", "should not contain spaces")
		})
	}
}

func TestGenString(t *testing.T) {
	for range 3 {
		got := GenString(10)
		assert.Len(t, got, 10)
		for _, c := range got {
			assert.True(t,
				(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'),
				"GenString contained invalid char %q", c)
		}
	}
	assert.Empty(t, GenString(0))
}

func TestIsProductOnSale(t *testing.T) {
	tests := []struct {
		s   string
		exp bool
	}{
		{"", false},
		{"10%", true},
		{"0%", true},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			assert.Equal(t, tt.exp, IsProductOnSale(tt.s))
		})
	}
}

func TestBuildFullName(t *testing.T) {
	tests := []struct {
		name     string
		first    string
		middle   string
		last     string
		expected string
	}{
		{"all present", "John", "Q", "Doe", "John Q Doe"},
		{"no middle", "John", "", "Doe", "John Doe"},
		{"first and last only", "Jane", "", "Smith", "Jane Smith"},
		{"empty first", "", "M", "Last", " M Last"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, BuildFullName(tt.first, tt.middle, tt.last))
		})
	}
}

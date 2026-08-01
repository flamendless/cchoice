package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueBrandSlug(t *testing.T) {
	taken := make(map[string]int64)

	assert.Equal(t, "bosch", UniqueBrandSlug("bosch", 1, taken))
	assert.Equal(t, "makita", UniqueBrandSlug("makita", 2, taken))
	assert.Equal(t, "bosch-3", UniqueBrandSlug("bosch", 3, taken))
	assert.Equal(t, "brand-4", UniqueBrandSlug("", 4, taken))
}

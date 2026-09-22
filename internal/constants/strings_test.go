package constants

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCDNExcludePrefixKey_includesProductImages(t *testing.T) {
	t.Parallel()
	assert.True(t, slices.Contains(CDNExcludePrefixKey, "product_images-"))
}

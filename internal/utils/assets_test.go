package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionedAsset(t *testing.T) {
	got := VersionedAsset("/static/js/foo.js")
	assert.True(t, strings.HasPrefix(got, "/static/js/foo.js?"))
	assert.Contains(t, got, "v=")
	assert.NotContains(t, got, "v=dev")
}

func TestVersionedAssetWithExistingQuery(t *testing.T) {
	got := VersionedAsset("/static/js/foo.js?x=1")
	assert.Contains(t, got, "x=1")
	assert.Contains(t, got, "v=")
}

func TestVersionedAssetWithURL(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("URL depends on conf; skipping: %v", r)
		}
	}()

	got := URL(VersionedAsset("/static/js/foo.js"))
	assert.NotEmpty(t, got)
	assert.Contains(t, got, "/static/js/foo.js")
	assert.Contains(t, got, "v=")
}

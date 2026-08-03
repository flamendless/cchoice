package datamigrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatGooseVersion(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "2026-01-09 16:43:36", FormatGooseVersion(20260109164336))
	assert.Equal(t, "2026-07-10 12:00:00", FormatGooseVersion(20260710120000))
	assert.Equal(t, "42", FormatGooseVersion(42))
}

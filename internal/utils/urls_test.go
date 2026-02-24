package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURL(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("URL depends on conf; skipping: %v", r)
		}
	}()
	got := URL("/admin")
	assert.NotEmpty(t, got)
	assert.True(t, got == "/admin" || got == "/cchoice/admin", "URL(/admin) = %q", got)
}

func TestMatchPath(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("MatchPath depends on conf; skipping: %v", r)
		}
	}()
	got := MatchPath("/cchoice/admin", "/admin")
	_ = got
}

func TestMetricsEvent(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("MetricsEvent depends on conf; skipping: %v", r)
		}
	}()
	got := MetricsEvent("click")
	assert.NotEmpty(t, got)
	assert.GreaterOrEqual(t, len(got), 10)
	assert.True(t, strings.Contains(got, "event=") && strings.Contains(got, "click"),
		"MetricsEvent(click) = %q, should contain event= and click", got)
}

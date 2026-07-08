package metrics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAllowedClientEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event string
		want  bool
	}{
		{name: "admin visit", event: EventAdminVisit, want: true},
		{name: "anon exec", event: EventAnonExec, want: true},
		{name: "unknown", event: "unknown_event", want: false},
		{name: "empty", event: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsAllowedClientEvent(tt.event))
		})
	}
}

func TestSanitizeClientEventValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event string
		value string
		want  string
	}{
		{
			name:  "strips newlines",
			event: EventAdminVisit,
			value: "hello\nworld",
			want:  "helloworld",
		},
		{
			name:  "truncates long value",
			event: EventAdminVisit,
			value: strings.Repeat("a", maxEventValueLen+10),
			want:  strings.Repeat("a", maxEventValueLen),
		},
		{
			name:  "truncates search query aggressively",
			event: EventSearchDesktop,
			value: strings.Repeat("q", highCardinalityValueLen+10),
			want:  strings.Repeat("q", highCardinalityValueLen),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, SanitizeClientEventValue(tt.event, tt.value))
		})
	}
}

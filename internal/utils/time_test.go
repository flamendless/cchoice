package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeToMinutes(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		wantMin int
		wantOk  bool
	}{
		{"empty", "", 0, false},
		{"datetime ISO", "2025-02-24 09:30:00", 9*60 + 30, true},
		{"datetime noon", "2025-01-01 12:00:00", 12 * 60, true},
		{"HHMMSS", "14:45:30", 14*60 + 45, true},
		{"HHMM", "08:00", 8 * 60, true},
		{"HHMM afternoon", "17:30", 17*60 + 30, true},
		{"midnight", "00:00", 0, true},
		{"invalid", "not-a-time", 0, false},
		{"bad date", "2025-13-01 10:00:00", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotOk := TimeToMinutes(tt.s)
			assert.Equal(t, tt.wantMin, gotMin)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func TestFormatDurationFromMinutes(t *testing.T) {
	tests := []struct {
		name   string
		m      int
		expect string
	}{
		{"zero", 0, "0h 0m"},
		{"one hour", 60, "1h 0m"},
		{"one hour thirty", 90, "1h 30m"},
		{"eight hours", 480, "8h 0m"},
		{"mixed", 125, "2h 5m"},
		{"negative", -1, "-"},
		{"negative large", -60, "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, FormatDurationFromMinutes(tt.m))
		})
	}
}

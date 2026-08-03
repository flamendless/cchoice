package utils

import (
	"testing"
	"time"

	"cchoice/internal/constants"

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

func TestConvertToPH(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{name: "empty", input: "", expect: ""},
		{name: "utc midnight to ph morning", input: "2026-08-03 00:00:00", expect: "2026-08-03 08:00:00"},
		{name: "utc afternoon to ph evening", input: "2026-08-03 10:30:00", expect: "2026-08-03 18:30:00"},
		{name: "invalid passthrough", input: "not-a-date", expect: "not-a-date"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, ConvertToPH(tt.input))
		})
	}
}

func TestFormatTimePH(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  time.Time
		expect string
	}{
		{name: "zero", input: time.Time{}, expect: ""},
		{name: "utc to ph", input: time.Date(2026, 8, 3, 5, 40, 0, 0, time.UTC), expect: "2026-08-03 13:40:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, FormatTimePH(tt.input))
		})
	}
}

func TestFormatTimePHUsesProjectLayout(t *testing.T) {
	t.Parallel()

	got := FormatTimePH(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	_, err := time.Parse(constants.DateTimeLayoutISO, got)
	assert.NoError(t, err)
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

package services

import (
	"testing"
	"time"

	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"github.com/stretchr/testify/assert"
)

func TestHolidayTypeParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected enums.HolidayType
	}{
		{
			name:     "regular holiday",
			input:    "regular",
			expected: enums.HOLIDAY_TYPE_REGULAR,
		},
		{
			name:     "regular uppercase",
			input:    "REGULAR",
			expected: enums.HOLIDAY_TYPE_REGULAR,
		},
		{
			name:     "special non-working",
			input:    "special_non_working",
			expected: enums.HOLIDAY_TYPE_SPECIAL_NON_WORKING,
		},
		{
			name:     "special non-working uppercase",
			input:    "SPECIAL_NON_WORKING",
			expected: enums.HOLIDAY_TYPE_SPECIAL_NON_WORKING,
		},
		{
			name:     "special working",
			input:    "special_working",
			expected: enums.HOLIDAY_TYPE_SPECIAL_WORKING,
		},
		{
			name:     "special working uppercase",
			input:    "SPECIAL_WORKING",
			expected: enums.HOLIDAY_TYPE_SPECIAL_WORKING,
		},
		{
			name:     "invalid type",
			input:    "invalid",
			expected: enums.HOLIDAY_TYPE_UNDEFINED,
		},
		{
			name:     "empty string",
			input:    "",
			expected: enums.HOLIDAY_TYPE_UNDEFINED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := enums.ParseHolidayTypeToEnum(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHolidayTypeString(t *testing.T) {
	tests := []struct {
		name     string
		input    enums.HolidayType
		expected string
	}{
		{
			name:     "regular",
			input:    enums.HOLIDAY_TYPE_REGULAR,
			expected: "REGULAR",
		},
		{
			name:     "special non-working",
			input:    enums.HOLIDAY_TYPE_SPECIAL_NON_WORKING,
			expected: "SPECIAL_NON_WORKING",
		},
		{
			name:     "special working",
			input:    enums.HOLIDAY_TYPE_SPECIAL_WORKING,
			expected: "SPECIAL_WORKING",
		},
		{
			name:     "undefined",
			input:    enums.HOLIDAY_TYPE_UNDEFINED,
			expected: "UNDEFINED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDateFormat(t *testing.T) {
	date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	expected := "2026-01-01"

	result := date.Format(constants.DateLayoutISO)

	assert.Equal(t, expected, result)
}

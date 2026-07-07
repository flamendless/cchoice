package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateOrderEarnedCPoints(t *testing.T) {
	tests := []struct {
		name        string
		totalAmount int64
		want        int64
	}{
		{"zero", 0, 0},
		{"negative", -100, 0},
		{"one hundred pesos", 10000, 10},
		{"ten thousand one eighty seven fifty pesos", 1018750, 1019},
		{"five thousand pesos", 500000, 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CalculateOrderEarnedCPoints(tt.totalAmount))
		})
	}
}

func TestFormatEarnedCPoints(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  string
	}{
		{"zero", 0, "-"},
		{"negative", -5, "-"},
		{"positive", 500, "500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatEarnedCPoints(tt.value))
		})
	}
}

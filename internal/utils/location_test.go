package utils

import (
	"math"
	"testing"
)

func TestHaversineDistanceMeters(t *testing.T) {
	tests := []struct {
		name      string
		lat1      float64
		lng1      float64
		lat2      float64
		lng2      float64
		expected  float64
		tolerance float64
	}{
		{
			name: "attendance sample coordinates",
			lat1: 14.333199079659577,
			lng1: 120.88151883134833,
			lat2: 14.3329,
			lng2: 120.8811,
			expected:  56.053545,
			tolerance: 0.5,
		},
		{
			name:      "zero distance",
			lat1:      14.333199079659577,
			lng1:      120.88151883134833,
			lat2:      14.333199079659577,
			lng2:      120.88151883134833,
			expected:  0,
			tolerance: 0.000001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HaversineDistanceMeters(
				tt.lat1,
				tt.lng1,
				tt.lat2,
				tt.lng2,
			)

			if math.Abs(got-tt.expected) > tt.tolerance {
				t.Fatalf(
					"distance mismatch: got=%.6f expected=%.6f tolerance=%.6f",
					got,
					tt.expected,
					tt.tolerance,
				)
			}
		})
	}
}

package utils

import (
	"cchoice/internal/database/queries"
	"database/sql"
	"testing"
)

func TestConvertWeightToKg(t *testing.T) {
	tests := []struct {
		name     string
		weight   sql.NullFloat64
		unit     sql.NullString
		expected float64
	}{
		{
			name:     "Kilograms",
			weight:   sql.NullFloat64{Float64: 2.5, Valid: true},
			unit:     sql.NullString{String: "kg", Valid: true},
			expected: 2.5,
		},
		{
			name:     "Grams to kg",
			weight:   sql.NullFloat64{Float64: 1500, Valid: true},
			unit:     sql.NullString{String: "g", Valid: true},
			expected: 1.5,
		},
		{
			name:     "Pounds to kg",
			weight:   sql.NullFloat64{Float64: 2.2, Valid: true},
			unit:     sql.NullString{String: "lb", Valid: true},
			expected: 0.9979024000000001,
		},
		{
			name:     "Ounces to kg",
			weight:   sql.NullFloat64{Float64: 35.274, Valid: true},
			unit:     sql.NullString{String: "oz", Valid: true},
			expected: 1.000000263,
		},
		{
			name:     "No unit specified",
			weight:   sql.NullFloat64{Float64: 3.0, Valid: true},
			unit:     sql.NullString{Valid: false},
			expected: 3.0,
		},
		{
			name:     "Invalid weight",
			weight:   sql.NullFloat64{Float64: -1.0, Valid: true},
			unit:     sql.NullString{String: "kg", Valid: true},
			expected: 1.0,
		},
		{
			name:     "No weight specified",
			weight:   sql.NullFloat64{Valid: false},
			unit:     sql.NullString{String: "kg", Valid: true},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertWeightToKg(tt.weight, tt.unit)
			if err != nil {
				t.Errorf("ConvertWeightToKg() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ConvertWeightToKg() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateTotalWeightFromCheckoutLines(t *testing.T) {
	tests := []struct {
		name          string
		checkoutLines []queries.GetCheckoutLinesByCheckoutIDRow
		expected      string
	}{
		{
			name: "Single item",
			checkoutLines: []queries.GetCheckoutLinesByCheckoutIDRow{
				{
					Weight:     sql.NullFloat64{Float64: 2.0, Valid: true},
					WeightUnit: sql.NullString{String: "kg", Valid: true},
					Quantity:   1,
				},
			},
			expected: "2.00",
		},
		{
			name: "Multiple items with different units",
			checkoutLines: []queries.GetCheckoutLinesByCheckoutIDRow{
				{
					Weight:     sql.NullFloat64{Float64: 1.5, Valid: true},
					WeightUnit: sql.NullString{String: "kg", Valid: true},
					Quantity:   2,
				},
				{
					Weight:     sql.NullFloat64{Float64: 500, Valid: true},
					WeightUnit: sql.NullString{String: "g", Valid: true},
					Quantity:   1,
				},
			},
			expected: "3.50",
		},
		{
			name:          "Empty cart",
			checkoutLines: []queries.GetCheckoutLinesByCheckoutIDRow{},
			expected:      "0.00",
		},
		{
			name: "Item with no weight specified",
			checkoutLines: []queries.GetCheckoutLinesByCheckoutIDRow{
				{
					Weight:     sql.NullFloat64{Valid: false},
					WeightUnit: sql.NullString{Valid: false},
					Quantity:   1,
				},
			},
			expected: "1.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateTotalWeightFromCheckoutLines(tt.checkoutLines)
			if err != nil {
				t.Errorf("CalculateTotalWeightFromCheckoutLines() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("CalculateTotalWeightFromCheckoutLines() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkConvertWeightToKg(b *testing.B) {
	testCases := []struct {
		name   string
		weight sql.NullFloat64
		unit   sql.NullString
	}{
		{
			name:   "Kilograms",
			weight: sql.NullFloat64{Float64: 2.5, Valid: true},
			unit:   sql.NullString{String: "kg", Valid: true},
		},
		{
			name:   "Grams",
			weight: sql.NullFloat64{Float64: 1500, Valid: true},
			unit:   sql.NullString{String: "g", Valid: true},
		},
		{
			name:   "Pounds",
			weight: sql.NullFloat64{Float64: 2.2, Valid: true},
			unit:   sql.NullString{String: "lb", Valid: true},
		},
		{
			name:   "Ounces",
			weight: sql.NullFloat64{Float64: 35.274, Valid: true},
			unit:   sql.NullString{String: "oz", Valid: true},
		},
		{
			name:   "NoUnit",
			weight: sql.NullFloat64{Float64: 3.0, Valid: true},
			unit:   sql.NullString{Valid: false},
		},
		{
			name:   "InvalidWeight",
			weight: sql.NullFloat64{Float64: -1.0, Valid: true},
			unit:   sql.NullString{String: "kg", Valid: true},
		},
		{
			name:   "NoWeight",
			weight: sql.NullFloat64{Valid: false},
			unit:   sql.NullString{String: "kg", Valid: true},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				if _, err := ConvertWeightToKg(tc.weight, tc.unit); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkCalculateTotalWeightFromCheckoutLines(b *testing.B) {
	smallCart := []queries.GetCheckoutLinesByCheckoutIDRow{
		{
			Weight:     sql.NullFloat64{Float64: 2.0, Valid: true},
			WeightUnit: sql.NullString{String: "kg", Valid: true},
			Quantity:   1,
		},
	}

	mediumCart := []queries.GetCheckoutLinesByCheckoutIDRow{
		{
			Weight:     sql.NullFloat64{Float64: 1.5, Valid: true},
			WeightUnit: sql.NullString{String: "kg", Valid: true},
			Quantity:   2,
		},
		{
			Weight:     sql.NullFloat64{Float64: 500, Valid: true},
			WeightUnit: sql.NullString{String: "g", Valid: true},
			Quantity:   1,
		},
		{
			Weight:     sql.NullFloat64{Float64: 2.2, Valid: true},
			WeightUnit: sql.NullString{String: "lb", Valid: true},
			Quantity:   1,
		},
	}

	largeCart := make([]queries.GetCheckoutLinesByCheckoutIDRow, 100)
	for i := range 100 {
		largeCart[i] = queries.GetCheckoutLinesByCheckoutIDRow{
			Weight:     sql.NullFloat64{Float64: float64(i%10 + 1), Valid: true},
			WeightUnit: sql.NullString{String: []string{"kg", "g", "lb", "oz"}[i%4], Valid: true},
			Quantity:   int64(i%5 + 1),
		}
	}

	emptyCart := []queries.GetCheckoutLinesByCheckoutIDRow{}

	benchmarks := []struct {
		name          string
		checkoutLines []queries.GetCheckoutLinesByCheckoutIDRow
	}{
		{"EmptyCart", emptyCart},
		{"SmallCart", smallCart},
		{"MediumCart", mediumCart},
		{"LargeCart", largeCart},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				if _, err := CalculateTotalWeightFromCheckoutLines(bm.checkoutLines); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkWeightCalculationPipeline(b *testing.B) {
	realisticCart := []queries.GetCheckoutLinesByCheckoutIDRow{
		{
			Weight:     sql.NullFloat64{Float64: 2.5, Valid: true},
			WeightUnit: sql.NullString{String: "kg", Valid: true},
			Quantity:   2,
		},
		{
			Weight:     sql.NullFloat64{Float64: 750, Valid: true},
			WeightUnit: sql.NullString{String: "g", Valid: true},
			Quantity:   3,
		},
		{
			Weight:     sql.NullFloat64{Float64: 1.2, Valid: true},
			WeightUnit: sql.NullString{String: "lb", Valid: true},
			Quantity:   1,
		},
		{
			Weight:     sql.NullFloat64{Float64: 16, Valid: true},
			WeightUnit: sql.NullString{String: "oz", Valid: true},
			Quantity:   2,
		},
		{
			Weight:     sql.NullFloat64{Valid: false},
			WeightUnit: sql.NullString{Valid: false},
			Quantity:   1,
		},
	}

	for b.Loop() {
		if _, err := CalculateTotalWeightFromCheckoutLines(realisticCart); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWeightConversionByUnit(b *testing.B) {
	units := []string{"kg", "g", "lb", "oz", ""}
	weight := 2.5

	for _, unit := range units {
		b.Run("Unit_"+unit, func(b *testing.B) {
			weightNull := sql.NullFloat64{Float64: weight, Valid: true}
			unitNull := sql.NullString{String: unit, Valid: unit != ""}

			b.ResetTimer()
			for b.Loop() {
				if _, err := ConvertWeightToKg(weightNull, unitNull); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

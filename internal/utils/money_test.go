package utils

import (
	"cchoice/internal/constants"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name     string
		price    int64
		currency string
	}{
		{"zero", 0, constants.PHP},
		{"positive", 10000, constants.PHP},
		{"large", 99999999, "USD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMoney(tt.price, tt.currency)
			require.NotNil(t, m)
			assert.Equal(t, tt.currency, m.Currency().Code)
		})
	}
}

func TestNewMoneyFromString(t *testing.T) {
	tests := []struct {
		name     string
		price    string
		currency string
		wantErr  bool
	}{
		{"valid integer", "100", constants.PHP, false},
		{"valid decimal", "99.50", constants.PHP, false},
		{"invalid", "not-a-number", constants.PHP, true},
		{"empty", "", constants.PHP, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMoneyFromString(tt.price, tt.currency)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, m)
			assert.Equal(t, tt.currency, m.Currency().Code)
		})
	}
}

func TestGetOrigAndDiscounted(t *testing.T) {
	tests := []struct {
		name                   string
		isOnSale               int64
		unitPriceWithVat       int64
		unitPriceWithVatCurr   string
		salePriceWithVat       sql.NullInt64
		salePriceWithVatCurr   sql.NullString
		wantDiscountPercentage string
	}{
		{
			name:                   "not on sale",
			isOnSale:               0,
			unitPriceWithVat:       10000,
			unitPriceWithVatCurr:   constants.PHP,
			salePriceWithVat:       sql.NullInt64{},
			salePriceWithVatCurr:   sql.NullString{},
			wantDiscountPercentage: "",
		},
		{
			name:                   "on sale 50% off",
			isOnSale:               1,
			unitPriceWithVat:       10000,
			unitPriceWithVatCurr:   constants.PHP,
			salePriceWithVat:       sql.NullInt64{Int64: 5000, Valid: true},
			salePriceWithVatCurr:   sql.NullString{String: constants.PHP, Valid: true},
			wantDiscountPercentage: "50%",
		},
		{
			name:                   "on sale 25% off",
			isOnSale:               1,
			unitPriceWithVat:       10000,
			unitPriceWithVatCurr:   constants.PHP,
			salePriceWithVat:       sql.NullInt64{Int64: 7500, Valid: true},
			salePriceWithVatCurr:   sql.NullString{String: constants.PHP, Valid: true},
			wantDiscountPercentage: "25%",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig, discounted, discountPct := GetOrigAndDiscounted(
				tt.isOnSale,
				tt.unitPriceWithVat,
				tt.unitPriceWithVatCurr,
				tt.salePriceWithVat,
				tt.salePriceWithVatCurr,
			)
			require.NotNil(t, orig)
			require.NotNil(t, discounted)
			assert.Equal(t, tt.wantDiscountPercentage, discountPct)
			if tt.isOnSale == 0 {
				assert.Same(t, orig, discounted)
			}
		})
	}
}

func TestGetDiscountAmount(t *testing.T) {
	tests := []struct {
		name             string
		isOnSale         int64
		unitPriceWithVat int64
		salePriceWithVat sql.NullInt64
		wantZero         bool
		wantAmount       int64
	}{
		{"not on sale", 0, 10000, sql.NullInt64{Int64: 5000, Valid: true}, true, 0},
		{"on sale", 1, 10000, sql.NullInt64{Int64: 7000, Valid: true}, false, 3000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := GetDiscountAmount(
				tt.isOnSale,
				tt.unitPriceWithVat,
				constants.PHP,
				tt.salePriceWithVat,
				sql.NullString{String: constants.PHP, Valid: true},
			)
			require.NotNil(t, m)
			assert.Equal(t, tt.wantAmount, m.Amount())
		})
	}
}

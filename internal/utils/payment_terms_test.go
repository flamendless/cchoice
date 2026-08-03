package utils

import (
	"testing"

	"cchoice/internal/enums"
	"cchoice/internal/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAmountInWordsPHP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		centavos int64
		want     string
	}{
		{"zero", 0, "Zero Pesos"},
		{"one peso", 100, "One Peso"},
		{"with centavos", 10150, "One Hundred One Pesos and Fifty Centavos"},
		{"large", 123456789, "One Million Two Hundred Thirty Four Thousand Five Hundred Sixty Seven Pesos and Eighty Nine Centavos"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, AmountInWordsPHP(tt.centavos))
		})
	}
}

func TestComputeDueDateFromTerms(t *testing.T) {
	t.Parallel()
	got, err := ComputeDueDateFromTerms("2026-08-01", 30, enums.PAYMENT_TERMS_DAYS)
	require.NoError(t, err)
	assert.Equal(t, "2026-09-01", got)

	got, err = ComputeDueDateFromTerms("2026-01-15", 30, enums.PAYMENT_TERMS_DAYS)
	require.NoError(t, err)
	assert.Equal(t, "2026-02-15", got)

	got, err = ComputeDueDateFromTerms("2026-08-01", 45, enums.PAYMENT_TERMS_DAYS)
	require.NoError(t, err)
	assert.Equal(t, "2026-09-16", got)

	got, err = ComputeDueDateFromTerms("2026-01-15", 2, enums.PAYMENT_TERMS_MONTHS)
	require.NoError(t, err)
	assert.Equal(t, "2026-03-15", got)

	got, err = ComputeDueDateFromTerms("2026-01-15", 1, enums.PAYMENT_TERMS_YEARS)
	require.NoError(t, err)
	assert.Equal(t, "2027-01-15", got)

	_, err = ComputeDueDateFromTerms("2026-01-15", 0, enums.PAYMENT_TERMS_DAYS)
	assert.ErrorIs(t, err, errs.ErrInvalidPaymentTerms)
}

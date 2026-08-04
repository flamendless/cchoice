package cmd

import (
	"database/sql"
	"testing"
	"time"

	"cchoice/internal/database/queries"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldRetainProductImage(t *testing.T) {
	t.Parallel()

	cutoff := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	old := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		row        queries.GetProductsWithImagesByBrandIDRow
		wantRetain bool
		wantLatest time.Time
	}{
		{
			name: "retain when updated_at is on or after cutoff",
			row: queries.GetProductsWithImagesByBrandIDRow{
				ImageCreatedAt: sql.NullTime{Time: old, Valid: true},
				ImageUpdatedAt: sql.NullTime{Time: newer, Valid: true},
			},
			wantRetain: true,
			wantLatest: newer,
		},
		{
			name: "update when both timestamps are before cutoff",
			row: queries.GetProductsWithImagesByBrandIDRow{
				ImageCreatedAt: sql.NullTime{Time: old, Valid: true},
				ImageUpdatedAt: sql.NullTime{Time: old, Valid: true},
			},
			wantRetain: false,
			wantLatest: old,
		},
		{
			name:       "update when no image timestamps",
			row:        queries.GetProductsWithImagesByBrandIDRow{},
			wantRetain: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotRetain, gotLatest := shouldRetainProductImage(tt.row, cutoff)
			assert.Equal(t, tt.wantRetain, gotRetain)
			if tt.wantLatest.IsZero() {
				assert.True(t, gotLatest.IsZero())
				return
			}
			require.Equal(t, tt.wantLatest, gotLatest)
		})
	}
}

func TestFindBatchImageColumnIndex(t *testing.T) {
	t.Parallel()

	idx, err := findBatchImageColumnIndex([]string{"Brand", "Product Code", "Name"}, "product code")
	require.NoError(t, err)
	assert.Equal(t, 1, idx)
}

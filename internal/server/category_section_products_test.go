package server_test

import (
	"strings"
	"testing"

	"cchoice/internal/constants"
	"cchoice/internal/errs"
	"cchoice/internal/server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCategoryProductBatchIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr error
	}{
		{
			name:  "single id",
			input: "abc123",
			want:  []string{"abc123"},
		},
		{
			name:  "multiple ids",
			input: "abc,def,ghi",
			want:  []string{"abc", "def", "ghi"},
		},
		{
			name:  "trims whitespace and dedupes",
			input: " abc , def , abc ",
			want:  []string{"abc", "def"},
		},
		{
			name:    "empty",
			input:   "  ",
			wantErr: errs.ErrInvalidParams,
		},
		{
			name:    "over batch limit",
			input:   "a,b,c,d,e",
			wantErr: errs.ErrInvalidParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := server.ParseCategoryProductBatchIDs(tt.input)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseCategoryProductBatchIDs_MaxBatchSize(t *testing.T) {
	t.Parallel()

	ids := make([]string, constants.DefaultShopBatchProductSections)
	for i := range ids {
		ids[i] = string(rune('a' + i))
	}
	input := strings.Join(ids, ",")

	got, err := server.ParseCategoryProductBatchIDs(input)
	require.NoError(t, err)
	assert.Len(t, got, constants.DefaultShopBatchProductSections)
}

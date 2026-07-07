package utils

import (
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetListingSortFromQuery(t *testing.T) {
	t.Parallel()

	sortBy, sortDir := GetListingSortFromQuery(url.Values{
		"sort_by":  {"created_at"},
		"sort_dir": {"asc"},
	})
	require.Equal(t, "CREATED_AT", sortBy)
	require.Equal(t, enums.LISTING_SORT_DIRECTION_ASC, sortDir)
}

func TestParseListingSortQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		query         url.Values
		allowedSortBy []string
		wantSortBy    string
		wantSortDir   enums.ListingSortDirection
		wantErr       error
	}{
		{
			name:          "accepts empty values",
			query:         url.Values{},
			allowedSortBy: []string{"CREATED_AT", "STATUS"},
			wantSortBy:    "",
			wantSortDir:   enums.LISTING_SORT_DIRECTION_UNDEFINED,
		},
		{
			name: "accepts allowed values",
			query: url.Values{
				"sort_by":  {"STATUS"},
				"sort_dir": {"DESC"},
			},
			allowedSortBy: []string{"CREATED_AT", "STATUS"},
			wantSortBy:    "STATUS",
			wantSortDir:   enums.LISTING_SORT_DIRECTION_DESC,
		},
		{
			name: "rejects invalid sort by",
			query: url.Values{
				"sort_by": {"UPDATED_AT"},
			},
			allowedSortBy: []string{"CREATED_AT", "STATUS"},
			wantSortBy:    "UPDATED_AT",
			wantErr:       errs.ErrEnumInvalid,
		},
		{
			name: "rejects invalid sort dir",
			query: url.Values{
				"sort_dir": {"INVALID"},
			},
			allowedSortBy: []string{"CREATED_AT", "STATUS"},
			wantSortDir:   enums.LISTING_SORT_DIRECTION_UNDEFINED,
			wantErr:       errs.ErrEnumInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sortBy, sortDir, err := ParseListingSortQuery(tt.query, tt.allowedSortBy...)
			require.Equal(t, tt.wantSortBy, sortBy)
			require.Equal(t, tt.wantSortDir, sortDir)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestNormalizeListingSort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		sortBy        string
		sortDir       enums.ListingSortDirection
		defaultSortBy string
		wantSortBy    string
		wantSortDir   enums.ListingSortDirection
	}{
		{
			name:          "defaults empty values",
			sortBy:        "",
			sortDir:       enums.LISTING_SORT_DIRECTION_UNDEFINED,
			defaultSortBy: "UPDATED_AT",
			wantSortBy:    "UPDATED_AT",
			wantSortDir:   enums.LISTING_SORT_DIRECTION_DESC,
		},
		{
			name:          "keeps explicit values",
			sortBy:        "CREATED_AT",
			sortDir:       enums.LISTING_SORT_DIRECTION_ASC,
			defaultSortBy: "UPDATED_AT",
			wantSortBy:    "CREATED_AT",
			wantSortDir:   enums.LISTING_SORT_DIRECTION_ASC,
		},
		{
			name:          "defaults only sort dir",
			sortBy:        "STATUS",
			sortDir:       enums.LISTING_SORT_DIRECTION_UNDEFINED,
			defaultSortBy: "CREATED_AT",
			wantSortBy:    "STATUS",
			wantSortDir:   enums.LISTING_SORT_DIRECTION_DESC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sortBy, sortDir := NormalizeListingSort(tt.sortBy, tt.sortDir, tt.defaultSortBy)
			require.Equal(t, tt.wantSortBy, sortBy)
			require.Equal(t, tt.wantSortDir, sortDir)
		})
	}
}

func TestGetLimit(t *testing.T) {
	type tc struct {
		input    string
		expected int64
		err      error
	}
	cases := []*tc{
		{
			input:    "",
			expected: 100,
			err:      nil,
		},
		{
			input:    "0",
			expected: 0,
			err:      errs.ErrInvalidParams,
		},
		{
			input:    "26",
			expected: 26,
			err:      nil,
		},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			res, err := GetLimit(c.input)
			if c.err != nil {
				require.Error(t, err)
			}
			require.Equal(t, c.expected, res)
		})
	}
}

func BenchmarkGetLimit(b *testing.B) {
	for i := range b.N {
		_, _ = GetLimit(strconv.Itoa(i))
	}
}

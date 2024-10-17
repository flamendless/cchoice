package utils

import (
	"cchoice/internal/errs"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

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
			err:      errs.ERR_INVALID_PARAMS,
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
	for i := 0; i < b.N; i++ {
		_, _ = GetLimit(strconv.Itoa(i))
	}
}

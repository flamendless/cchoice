package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSumHoursFromSQLCValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    float64
		wantErr bool
	}{
		{name: "float64", input: float64(12.5), want: 12.5},
		{name: "int64", input: int64(8), want: 8},
		{name: "invalid", input: "nope", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := sumHoursFromSQLCValue(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

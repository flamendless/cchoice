package utils

import (
	"testing"

	"cchoice/internal/errs"

	"github.com/stretchr/testify/assert"
)

func TestValidateExternalURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "valid https", input: "https://www.lazada.com.ph/products/test", wantErr: nil},
		{name: "valid http", input: "http://shopee.ph/product/123", wantErr: nil},
		{name: "empty", input: "", wantErr: errs.ErrInvalidFormat},
		{name: "no scheme", input: "www.lazada.com.ph", wantErr: errs.ErrInvalidFormat},
		{name: "invalid scheme", input: "ftp://example.com", wantErr: errs.ErrInvalidFormat},
		{name: "whitespace only", input: "   ", wantErr: errs.ErrInvalidFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateExternalURL(tt.input)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

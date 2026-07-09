package forms_test

import (
	"testing"

	"cchoice/internal/errs"
	"cchoice/internal/server/forms"

	"github.com/stretchr/testify/assert"
)

func TestShippingAddressQuery_Validate(t *testing.T) {
	tests := []struct {
		name    string
		query   forms.ShippingAddressQuery
		wantErr error
	}{
		{
			name:  "provinces no extra fields",
			query: forms.ShippingAddressQuery{Data: "provinces"},
		},
		{
			name:  "cities with province",
			query: forms.ShippingAddressQuery{Data: "cities", Province: "Metro Manila"},
		},
		{
			name:    "cities missing province",
			query:   forms.ShippingAddressQuery{Data: "cities"},
			wantErr: errs.ErrInvalidParams,
		},
		{
			name:  "barangays with city",
			query: forms.ShippingAddressQuery{Data: "barangays", City: "Quezon City"},
		},
		{
			name:    "barangays missing city",
			query:   forms.ShippingAddressQuery{Data: "barangays"},
			wantErr: errs.ErrInvalidParams,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShippingQuotationForm_Validate(t *testing.T) {
	tests := []struct {
		name    string
		form    forms.ShippingQuotationForm
		wantErr error
	}{
		{
			name: "ncr without city or barangay",
			form: forms.ShippingQuotationForm{Province: "National Capital Region (NCR)"},
		},
		{
			name: "complete non-ncr",
			form: forms.ShippingQuotationForm{
				Province: "Bulacan",
				City:     "Malolos",
				Barangay: "Sample",
			},
		},
		{
			name:    "non-ncr missing barangay",
			form:    forms.ShippingQuotationForm{Province: "Bulacan", City: "Malolos"},
			wantErr: errs.ErrInvalidParams,
		},
		{
			name:    "non-ncr missing province",
			form:    forms.ShippingQuotationForm{City: "Malolos", Barangay: "Sample"},
			wantErr: errs.ErrInvalidParams,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.form.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

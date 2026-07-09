package forms_test

import (
	"testing"

	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/server/forms"

	"github.com/stretchr/testify/assert"
)

func TestCustomerRegisterForm_Normalize(t *testing.T) {
	form := forms.CustomerRegisterForm{MobileNo: "9171234567"}
	form.Normalize()
	assert.Equal(t, "+639171234567", form.MobileNo)

	alreadyPrefixed := forms.CustomerRegisterForm{MobileNo: "+639171234567"}
	alreadyPrefixed.Normalize()
	assert.Equal(t, "+639171234567", alreadyPrefixed.MobileNo)
}

func TestCustomerRegisterForm_Validate(t *testing.T) {
	valid := forms.CustomerRegisterForm{
		MobileNo:     "+639171234567",
		CustomerType: enums.CUSTOMER_TYPE_CUSTOMER.String(),
	}

	tests := []struct {
		name    string
		form    forms.CustomerRegisterForm
		wantErr bool
		errMsg  string
	}{
		{name: "valid customer", form: valid},
		{
			name: "valid company",
			form: forms.CustomerRegisterForm{
				MobileNo:     "+639171234567",
				CustomerType: enums.CUSTOMER_TYPE_COMPANY.String(),
				CompanyName:  "Acme Corp",
			},
		},
		{
			name: "invalid mobile",
			form: forms.CustomerRegisterForm{
				MobileNo:     "09171234567",
				CustomerType: enums.CUSTOMER_TYPE_CUSTOMER.String(),
			},
			wantErr: true,
		},
		{
			name: "company missing name",
			form: forms.CustomerRegisterForm{
				MobileNo:     "+639171234567",
				CustomerType: enums.CUSTOMER_TYPE_COMPANY.String(),
			},
			wantErr: true,
			errMsg:  "company name is required for company accounts",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.form.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
				if tt.name == "invalid mobile" {
					assert.ErrorIs(t, err, errs.ErrValidationInvalidMobileNumber)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestCustomerProfileUpdateForm_NormalizeAndValidate(t *testing.T) {
	form := forms.CustomerProfileUpdateForm{MobileNo: "9171234567"}
	form.Normalize()
	assert.Equal(t, "+639171234567", form.MobileNo)
	assert.NoError(t, form.Validate())

	invalid := forms.CustomerProfileUpdateForm{MobileNo: "invalid"}
	invalid.Normalize()
	assert.ErrorIs(t, invalid.Validate(), errs.ErrValidationInvalidMobileNumber)
}

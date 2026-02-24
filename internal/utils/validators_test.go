package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateNotBlank(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		key     string
		wantErr bool
	}{
		{"non-empty", "hello", "field", false},
		{"blank", "", "field", true},
		{"whitespace only", "   ", "field", true},
		{"single char", "a", "key", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotBlank(tt.field, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid 8", "12345678", false},
		{"valid 32", "12345678901234567890123456789012", false},
		{"valid middle", "user_name_1", false},
		{"too short", "short", true},
		{"too long", "thisusernameiswaytoolongandexceedsthirtytwocharacters", true},
		{"blank", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePW(t *testing.T) {
	tests := []struct {
		name    string
		pw      string
		wantErr bool
	}{
		{"valid 8", "password", false},
		{"valid 32", "12345678901234567890123456789012", false},
		{"too short", "short", true},
		{"too long", "thispasswordiswaytoolongandexceedsthirtytwocharacters", true},
		{"blank", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePW(tt.pw)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUserReg(t *testing.T) {
	validMobile := "+639171234567"
	validEmail := "user@example.com"
	validPW := "password8"

	tests := []struct {
		name    string
		data    ValidateUserRegInput
		wantErr bool
	}{
		{
			name: "valid",
			data: ValidateUserRegInput{
				FirstName:       "John",
				MiddleName:      "M",
				LastName:        "Doe",
				Email:           validEmail,
				Password:        validPW,
				ConfirmPassword: validPW,
				MobileNo:        validMobile,
			},
			wantErr: false,
		},
		{
			name: "blank first name",
			data: ValidateUserRegInput{
				FirstName:       "",
				MiddleName:      "M",
				LastName:        "Doe",
				Email:           validEmail,
				Password:        validPW,
				ConfirmPassword: validPW,
				MobileNo:        validMobile,
			},
			wantErr: true,
		},
		{
			name: "password mismatch",
			data: ValidateUserRegInput{
				FirstName:       "John",
				MiddleName:      "M",
				LastName:        "Doe",
				Email:           validEmail,
				Password:        validPW,
				ConfirmPassword: "different8",
				MobileNo:        validMobile,
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			data: ValidateUserRegInput{
				FirstName:       "John",
				MiddleName:      "M",
				LastName:        "Doe",
				Email:           "not-an-email",
				Password:        validPW,
				ConfirmPassword: validPW,
				MobileNo:        validMobile,
			},
			wantErr: true,
		},
		{
			name: "invalid mobile prefix",
			data: ValidateUserRegInput{
				FirstName:       "John",
				MiddleName:      "M",
				LastName:        "Doe",
				Email:           validEmail,
				Password:        validPW,
				ConfirmPassword: validPW,
				MobileNo:        "+1234567890123",
			},
			wantErr: true,
		},
		{
			name: "mobile too short",
			data: ValidateUserRegInput{
				FirstName:       "John",
				MiddleName:      "M",
				LastName:        "Doe",
				Email:           validEmail,
				Password:        validPW,
				ConfirmPassword: validPW,
				MobileNo:        "+63917",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserReg(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

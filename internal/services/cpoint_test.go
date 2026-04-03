package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCode(t *testing.T) {
	svc := &CPointService{}

	tests := []struct {
		name string
	}{
		{"generates valid code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := svc.GenerateCode()

			assert.Len(t, code, 14, "code should be 14 characters")
			assert.True(t, strings.HasPrefix(code, "CP-"), "code should start with 'CP-'")

			parts := strings.Split(code, "-")
			assert.Len(t, parts, 4, "code should have 4 parts")
			assert.Equal(t, "CP", parts[0], "first part should be 'CP'")

			validChars := "ABCDEFGHJKMNPQRSTUVWXYZ123456789"
			for i := 1; i <= 3; i++ {
				for _, c := range parts[i] {
					assert.True(t, strings.Contains(validChars, string(c)), "character %c should be in valid chars", c)
				}
			}
		})
	}
}

func TestGenerateCode_Uniqueness(t *testing.T) {
	svc := &CPointService{}
	codes := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		code := svc.GenerateCode()
		assert.False(t, codes[code], "code %s should be unique", code)
		codes[code] = true
	}
}

func TestValidateCode(t *testing.T) {
	svc := &CPointService{}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"valid code", "CP-ABC-123-DEF", false},
		{"valid code with numbers 2-9", "CP-A2B-3C4-D5E", false},
		{"valid code max values", "CP-XYZ-999-ZXY", false},
		{"invalid prefix", "XX-ABC-123-DEF", true},
		{"invalid length short", "CP-AB-12-DE", true},
		{"invalid length long", "CP-ABCD-123-DEF", true},
		{"invalid char O in code", "CP-AOC-123-DEF", true},
		{"invalid char 0 in code", "CP-A0C-123-DEF", true},
		{"invalid char I in code", "CP-ABI-123-DEF", true},
		{"invalid char L in code", "CP-ABL-123-DEF", true},
		{"missing hyphens", "CPABC123DEF", true},
		{"too many hyphens", "CP-ABC-12-3-DEF", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateCode(tt.code)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func BenchmarkGenerateCode(b *testing.B) {
	svc := &CPointService{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GenerateCode()
	}
}

func BenchmarkValidateCode_Valid(b *testing.B) {
	svc := &CPointService{}
	code := "CP-ABC-123-DEF"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.ValidateCode(code)
	}
}

func BenchmarkValidateCode_Invalid(b *testing.B) {
	svc := &CPointService{}
	code := "XX-ABC-123-DEF"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.ValidateCode(code)
	}
}

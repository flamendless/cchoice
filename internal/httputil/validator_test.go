package httputil_test

import (
	"testing"

	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubEncoder struct {
	decoded map[string]int64
}

func (s stubEncoder) Name() string { return "stub" }

func (s stubEncoder) Encode(id int64) string { return "" }

func (s stubEncoder) Decode(str string) int64 {
	if id, ok := s.decoded[str]; ok {
		return id
	}
	return encode.INVALID
}

func TestRequireEncodedID(t *testing.T) {
	enc := stubEncoder{decoded: map[string]int64{"valid-id": 42}}

	tests := []struct {
		name    string
		id      string
		wantID  string
		wantErr error
	}{
		{name: "valid", id: "valid-id", wantID: "valid-id"},
		{name: "invalid", id: "bad-id", wantErr: errs.ErrInvalidParams},
		{name: "empty", id: "", wantErr: errs.ErrInvalidParams},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := httputil.RequireEncodedID(enc, tt.id)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, gotID)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, gotID)
		})
	}
}

func TestValidator_CustomTags(t *testing.T) {
	type payload struct {
		Mobile      string `validate:"ph_mobile"`
		Email       string `validate:"ph_email"`
		Password    string `validate:"ph_password"`
		Search      string `validate:"min_search"`
		UserType    string `validate:"user_type"`
		BrandStatus string `validate:"brand_status"`
	}

	tests := []struct {
		name    string
		data    payload
		wantErr bool
	}{
		{
			name: "all valid",
			data: payload{
				Mobile:      "+639171234567",
				Email:       "user@example.com",
				Password:    "Password123-_.?#@",
				Search:      "abc",
				UserType:    enums.USER_TYPE_CUSTOMER.String(),
				BrandStatus: enums.BRAND_STATUS_ACTIVE.String(),
			},
		},
		{
			name: "invalid mobile prefix",
			data: payload{
				Mobile:   "09171234567",
				Email:    "user@example.com",
				Password: "Password123-_.?#@",
				Search:   "abc",
				UserType: enums.USER_TYPE_CUSTOMER.String(),
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			data: payload{
				Mobile:   "+639171234567",
				Email:    "not-an-email",
				Password: "Password123-_.?#@",
				Search:   "abc",
				UserType: enums.USER_TYPE_CUSTOMER.String(),
			},
			wantErr: true,
		},
		{
			name: "search too short",
			data: payload{
				Mobile:   "+639171234567",
				Email:    "user@example.com",
				Password: "Password123-_.?#@",
				Search:   "ab",
				UserType: enums.USER_TYPE_CUSTOMER.String(),
			},
			wantErr: true,
		},
		{
			name: "invalid user type",
			data: payload{
				Mobile:   "+639171234567",
				Email:    "user@example.com",
				Password: "Password123-_.?#@",
				Search:   "abc",
				UserType: "INVALID",
			},
			wantErr: true,
		},
		{
			name: "empty brand status allowed",
			data: payload{
				Mobile:   "+639171234567",
				Email:    "user@example.com",
				Password: "Password123-_.?#@",
				Search:   "abc",
				UserType: enums.USER_TYPE_CUSTOMER.String(),
			},
		},
		{
			name: "invalid brand status",
			data: payload{
				Mobile:      "+639171234567",
				Email:       "user@example.com",
				Password:    "Password123-_.?#@",
				Search:      "abc",
				UserType:    enums.USER_TYPE_CUSTOMER.String(),
				BrandStatus: "NOT_A_STATUS",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := httputil.Validator().Struct(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

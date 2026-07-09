package httputil_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/server/forms"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestBindQuery_ShippingAddress(t *testing.T) {
	tests := []struct {
		name    string
		query   url.Values
		wantErr bool
	}{
		{
			name:    "provinces ok",
			query:   url.Values{"data": {"provinces"}},
			wantErr: false,
		},
		{
			name:    "cities missing province",
			query:   url.Values{"data": {"cities"}},
			wantErr: true,
		},
		{
			name:    "cities with province",
			query:   url.Values{"data": {"cities"}, "province": {"Metro Manila"}},
			wantErr: false,
		},
		{
			name:    "barangays missing city",
			query:   url.Values{"data": {"barangays"}},
			wantErr: true,
		},
		{
			name:    "invalid data",
			query:   url.Values{"data": {"invalid"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/shipping/address?"+tt.query.Encode(), nil)
			var q forms.ShippingAddressQuery
			err := httputil.BindQuery(req, &q)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBindForm_ShippingQuotation(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "ncr skips city barangay",
			body:    "province=National+Capital+Region+%28NCR%29",
			wantErr: false,
		},
		{
			name:    "non-ncr missing fields",
			body:    "province=Bulacan",
			wantErr: true,
		},
		{
			name: "non-ncr complete",
			body: strings.Join([]string{
				"province=Bulacan",
				"city=Malolos",
				"barangay=Sample",
			}, "&"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/shipping/quotation", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			var f forms.ShippingQuotationForm
			err := httputil.BindForm(req, &f)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBindQuery_SearchMinLength(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/search/products?q=ab", nil)
	var q forms.SearchProductsQuery
	err := httputil.BindQuery(req, &q)
	assert.Error(t, err)
}

func TestTrimStrings(t *testing.T) {
	type payload struct {
		Name string `form:"name"`
	}
	p := payload{Name: "  hello  "}
	httputil.TrimStrings(&p)
	assert.Equal(t, "hello", p.Name)
}

func TestErrorMessage_InvalidParams(t *testing.T) {
	msg := httputil.ErrorMessage(errs.ErrInvalidParams)
	assert.Equal(t, errs.ErrInvalidParams.Error(), msg)
}

func TestCustomerRegisterNormalize(t *testing.T) {
	body := strings.Join([]string{
		"first_name=John",
		"middle_name=M",
		"last_name=Doe",
		"birthdate=2000-01-01",
		"sex=MALE",
		"email=john@example.com",
		"mobile_no=9171234567",
		"password=Password123-_.?#@",
		"confirm_password=Password123-_.?#@",
		"customer_type=CUSTOMER",
	}, "&")
	req := httptest.NewRequest(http.MethodPost, "/customer/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var f forms.CustomerRegisterForm
	err := httputil.BindPostForm(req, &f)
	require.NoError(t, err)
	assert.Equal(t, "+639171234567", f.MobileNo)
}

func TestBindPath_ProductSlug(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/product/my-slug", nil)
	req = withChiParam(req, "slug", "my-slug")

	var pathReq forms.ProductSlugPath
	err := httputil.BindPath(req, &pathReq)
	require.NoError(t, err)
	assert.Equal(t, "my-slug", pathReq.Slug)
}

func TestBindPath_MissingSlug(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/product/", nil)
	req = withChiParam(req, "slug", "")

	var pathReq forms.ProductSlugPath
	err := httputil.BindPath(req, &pathReq)
	assert.Error(t, err)
}

func TestBindPath_NonPointerTarget(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/product/foo", nil)
	var pathReq forms.ProductSlugPath
	err := httputil.BindPath(req, pathReq)
	assert.ErrorIs(t, err, errs.ErrValidationTargetMustBeAPointer)
}

func TestBindJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name" validate:"required"`
	}
	body := `{"name":" widget "}`
	req := httptest.NewRequest(http.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var p payload
	err := httputil.BindJSON(req, &p)
	require.NoError(t, err)
	assert.Equal(t, "widget", p.Name)
}

func TestBindJSON_InvalidBody(t *testing.T) {
	type payload struct {
		Name string `json:"name" validate:"required"`
	}
	req := httptest.NewRequest(http.MethodPost, "/api", strings.NewReader(`{invalid`))

	var p payload
	err := httputil.BindJSON(req, &p)
	assert.ErrorIs(t, err, errs.ErrJSONDecode)
}

func TestBindPostForm_OnlyPostValues(t *testing.T) {
	body := "email=staff@example.com&user_type=STAFF"
	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password?email=ignored@example.com", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var f forms.ForgotPasswordForm
	err := httputil.BindPostForm(req, &f)
	require.NoError(t, err)
	assert.Equal(t, "staff@example.com", f.Email)
	assert.Equal(t, "STAFF", f.UserType)
}

func TestBindQuery_TrimsWhitespace(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/search/products?q=%20abc%20", nil)
	var q forms.SearchProductsQuery
	err := httputil.BindQuery(req, &q)
	require.NoError(t, err)
	assert.Equal(t, "abc", q.Q)
}

func TestErrorMessage_ValidationTags(t *testing.T) {
	tests := []struct {
		name    string
		query   url.Values
		contain string
	}{
		{
			name:    "required",
			query:   url.Values{},
			contain: "is required",
		},
		{
			name:    "oneof",
			query:   url.Values{"data": {"bad"}},
			contain: "invalid value",
		},
		{
			name:    "eqfield",
			query:   url.Values{"token": {"abc"}, "new_password": {"Password123-_.?#@"}, "confirm_password": {"different"}},
			contain: "Passwords must match",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?"+tt.query.Encode(), nil)
			var q forms.ShippingAddressQuery
			if tt.name == "eqfield" {
				body := tt.query.Encode()
				req = httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				var f forms.ResetPasswordForm
				err := httputil.BindPostForm(req, &f)
				require.Error(t, err)
				assert.Contains(t, httputil.ErrorMessage(err), tt.contain)
				return
			}
			err := httputil.BindQuery(req, &q)
			require.Error(t, err)
			assert.Contains(t, httputil.ErrorMessage(err), tt.contain)
		})
	}
}

func TestPageOrDefault(t *testing.T) {
	tests := []struct {
		name    string
		page    int
		def     int
		want    int
	}{
		{name: "zero", page: 0, def: 1, want: 1},
		{name: "negative", page: -2, def: 1, want: 1},
		{name: "positive", page: 3, def: 1, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, httputil.PageOrDefault(tt.page, tt.def))
		})
	}
}

func TestBindMultipartForm_NilForm(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	var f forms.AdminBrandCreateForm
	err := httputil.BindMultipartForm(req, &f)
	assert.ErrorIs(t, err, errs.ErrInvalidParams)
}

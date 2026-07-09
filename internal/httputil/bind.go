package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"cchoice/internal/errs"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
)

var decoder = form.NewDecoder()

func bindAndValidate(dst any, decode func() error) error {
	if err := decode(); err != nil {
		return err
	}
	TrimStrings(dst)
	if n, ok := dst.(interface{ Normalize() }); ok {
		n.Normalize()
	}
	return validateStruct(dst)
}

func bindValues(dst any, values url.Values) error {
	return bindAndValidate(dst, func() error {
		if err := decoder.Decode(dst, values); err != nil {
			return fmt.Errorf("%w: %w", errs.ErrInvalidParams, err)
		}
		return nil
	})
}

func BindQuery(r *http.Request, dst any) error {
	return bindValues(dst, r.URL.Query())
}

func BindForm(r *http.Request, dst any) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrInvalidParams, err)
	}
	return bindValues(dst, r.Form)
}

func BindPostForm(r *http.Request, dst any) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrInvalidParams, err)
	}
	return bindValues(dst, r.PostForm)
}

func BindPath(r *http.Request, dst any) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return errs.ErrValidationTargetMustBeAPointer
	}

	rt := rv.Elem().Type()
	for i := range rt.NumField() {
		field := rt.Field(i)
		paramName := field.Tag.Get("param")
		if paramName == "" {
			continue
		}
		rv.Elem().Field(i).SetString(chi.URLParam(r, paramName))
	}

	return bindAndValidate(dst, func() error { return nil })
}

func BindJSON(r *http.Request, dst any) error {
	return bindAndValidate(dst, func() error {
		if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
			return fmt.Errorf("%w: %w", errs.ErrJSONDecode, err)
		}
		return nil
	})
}

func BindMultipartForm(r *http.Request, dst any) error {
	if r.MultipartForm == nil {
		return errs.ErrInvalidParams
	}
	return bindValues(dst, r.MultipartForm.Value)
}

func TrimStrings(dst any) {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer {
		return
	}
	trimValue(rv.Elem())
}

func trimValue(v reflect.Value) {
	switch v.Kind() {
	case reflect.String:
		if v.CanSet() {
			v.SetString(strings.TrimSpace(v.String()))
		}
	case reflect.Pointer:
		if v.IsNil() {
			return
		}
		trimValue(v.Elem())
	case reflect.Struct:
		for _, field := range v.Fields() {
			if field.CanSet() {
				trimValue(field)
			}
		}
	}
}

func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	verrs, ok := errors.AsType[validator.ValidationErrors](err)
	if ok && len(verrs) > 0 {
		fe := verrs[0]
		field := fe.Field()
		switch fe.Tag() {
		case "required", "required_if", "required_unless":
			return field + " is required"
		case "oneof":
			return field + " has an invalid value"
		case "ph_mobile":
			return errs.ErrValidationInvalidMobileNumber.Error()
		case "ph_email", "email":
			return "Invalid email format"
		case "ph_password":
			return "Invalid password format"
		case "min_search":
			return errs.ErrInvalidParams.Error()
		case "eqfield":
			return "Passwords must match"
		case "user_type", "brand_status":
			return errs.ErrInvalidParams.Error()
		default:
			return field + " is invalid"
		}
	}

	if errors.Is(err, errs.ErrInvalidParams) {
		return errs.ErrInvalidParams.Error()
	}

	return err.Error()
}

func ReadBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

func PageOrDefault(page int, defaultPage int) int {
	if page <= 0 {
		return defaultPage
	}
	return page
}

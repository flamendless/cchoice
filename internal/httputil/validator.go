package httputil

import (
	"reflect"
	"strings"
	"sync"

	"cchoice/internal/constants"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"

	"github.com/go-playground/validator/v10"
)

var (
	validate     *validator.Validate
	validateOnce sync.Once
)

func Validator() *validator.Validate {
	validateOnce.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			if name := fld.Tag.Get("form"); name != "" {
				return name
			}
			if name := fld.Tag.Get("param"); name != "" {
				return name
			}
			if name := fld.Tag.Get("json"); name != "" {
				return strings.SplitN(name, ",", 2)[0]
			}
			return fld.Name
		})

		_ = validate.RegisterValidation("ph_mobile", validatePHMobile)
		_ = validate.RegisterValidation("ph_email", validatePHEmail)
		_ = validate.RegisterValidation("ph_password", validatePHPassword)
		_ = validate.RegisterValidation("min_search", validateMinSearch)
		_ = validate.RegisterValidation("user_type", validateUserType)
		_ = validate.RegisterValidation("brand_status", validateBrandStatus)
	})
	return validate
}

func validatePHMobile(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	return strings.HasPrefix(v, constants.PHMobilePrefix) && len(v) == 13
}

func validatePHEmail(fl validator.FieldLevel) bool {
	return constants.ReEmail.MatchString(fl.Field().String())
}

func validatePHPassword(fl validator.FieldLevel) bool {
	return constants.RePassword.MatchString(fl.Field().String())
}

func validateMinSearch(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) >= constants.MinSearchQueryLength
}

func validateUserType(fl validator.FieldLevel) bool {
	return enums.ParseUserTypeToEnum(fl.Field().String()).IsValid()
}

func validateBrandStatus(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	if v == "" {
		return true
	}
	return enums.ParseBrandStatusToEnum(v) != enums.BRAND_STATUS_UNDEFINED
}

func validateStruct(dst any) error {
	if err := Validator().Struct(dst); err != nil {
		return err
	}
	if v, ok := dst.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func RequireEncodedID(enc encode.IEncode, id string) (string, error) {
	if enc.Decode(id) == encode.INVALID {
		return "", errs.ErrInvalidParams
	}
	return id, nil
}

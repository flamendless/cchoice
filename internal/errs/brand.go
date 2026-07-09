package errs

import "errors"

var (
	ErrBrand                  = errors.New("[BRAND]: Error on brand service")
	ErrBrandNotFound          = errors.New("[BRAND]: Brand not found")
	ErrBrandAlreadyExist      = errors.New("[BRAND]: Brand already exists")
	ErrBrandIDAndNameRequired = errors.New("[BRAND]: id and name are required")
	ErrBrandLogoRequired      = errors.New("[BRAND]: Logo image is required")
	ErrBrandNameRequired      = errors.New("[BRAND]: Brand name is required")
	ErrBrandDeleteFailed      = errors.New("[BRAND]: Failed to delete brand")
	ErrBrandUpdateFailed      = errors.New("[BRAND]: Failed to update brand")
)

package errs

import "errors"

var (
	ErrBrand             = errors.New("[BRAND]: Error on brand service")
	ErrBrandNotFound     = errors.New("[BRAND]: Brand not found")
	ErrBrandAlreadyExist = errors.New("[BRAND]: Brand already exists")
)

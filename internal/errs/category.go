package errs

import "errors"

var (
	ErrCategory              = errors.New("[CATEGORY]: Error on category service")
	ErrCategoryAlreadyExists = errors.New("[CATEGORY]: Category already exists")
	ErrCategoryNotFound      = errors.New("[CATEGORY]: Category not found")
	ErrCategoryPairExists    = errors.New("[CATEGORY]: Category and subcategory pair already exists")
)

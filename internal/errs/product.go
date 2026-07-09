package errs

import "errors"

var (
	ErrProduct                    = errors.New("[PRODUCT]: Error on product service")
	ErrProductBrandRequired       = errors.New("[PRODUCT]: Brand required")
	ErrProductInvalidID           = errors.New("[PRODUCT]: Invalid product ID")
	ErrProductAllFieldsRequired   = errors.New("[PRODUCT]: All fields are required")
	ErrProductInvalidBrand        = errors.New("[PRODUCT]: Invalid brand")
	ErrProductCategoryRequired    = errors.New("[PRODUCT]: Category and subcategory are required")
	ErrProductSerialExists        = errors.New("[PRODUCT]: Serial already exists")
	ErrProductSerialValidateFailed = errors.New("[PRODUCT]: Failed to validate serial")
	ErrProductInvalidPrice        = errors.New("[PRODUCT]: Invalid price")
	ErrProductSpecsRequired       = errors.New("[PRODUCT]: All product specs are required")
	ErrProductInvalidStocksIn     = errors.New("[PRODUCT]: Invalid stocks in")
	ErrProductInvalidSalePrice    = errors.New("[PRODUCT]: Invalid sale price")
	ErrProductSaleDatesRequired   = errors.New("[PRODUCT]: Sale start and end dates are required when sale price is set")
	ErrProductImageRequired       = errors.New("[PRODUCT]: Product image is required")
	ErrProductDraftOnly           = errors.New("[PRODUCT]: Only draft products can be edited")
	ErrProductStatusRequired      = errors.New("[PRODUCT]: Status is required")
	ErrProductInvalidStatus       = errors.New("[PRODUCT]: Invalid status")
	ErrProductUpdateStatusFailed  = errors.New("[PRODUCT]: Failed to update product status")
	ErrProductDeleteFailed        = errors.New("[PRODUCT]: Failed to delete product")
	ErrProductImageUploadFailed   = errors.New("[PRODUCT]: Failed to upload image")
)

package errs

import "errors"

var (
	ErrPromo              = errors.New("[PROMO]: Error on promo service")
	ErrPromoNotFound      = errors.New("[PROMO]: Promo not found")
	ErrPromoGetFailed     = errors.New("[PROMO]: Failed to get promo")
	ErrPromoDeleteFailed  = errors.New("[PROMO]: Failed to delete promo")
	ErrPromoMediaURLRequired = errors.New("[PROMO]: Media URL is required for video type")
	ErrPromoMediaFileRequired = errors.New("[PROMO]: Media file is required for image type")
)

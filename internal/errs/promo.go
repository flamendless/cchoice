package errs

import "errors"

var (
	ErrPromo                  = errors.New("[PROMO]: Error on promo service")
	ErrPromoNotFound          = errors.New("[PROMO]: Promo not found")
	ErrPromoGetFailed         = errors.New("[PROMO]: Failed to get promo")
	ErrPromoDeleteFailed      = errors.New("[PROMO]: Failed to delete promo")
	ErrPromoMediaURLRequired  = errors.New("[PROMO]: Media URL is required for video type")
	ErrPromoMediaFileRequired = errors.New("[PROMO]: Media file is required for image type")
	ErrPromoLinkConflict      = errors.New("[PROMO]: Only one link type can be set")
	ErrPromoLinkRequired      = errors.New("[PROMO]: Link is required for the selected link type")
	ErrPromoLinkInvalid       = errors.New("[PROMO]: Link URL must be an absolute URL or a path starting with /")
	ErrPromoLinkImageOnly     = errors.New("[PROMO]: Links are only supported for image banners")
)

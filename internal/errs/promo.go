package errs

import "errors"

var (
	ErrPromo         = errors.New("[PROMO]: Error on promo service")
	ErrPromoNotFound = errors.New("[PROMO]: Promo not found")
)

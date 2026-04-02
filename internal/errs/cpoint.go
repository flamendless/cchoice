package errs

import "errors"

var (
	ErrCpointNotFound           = errors.New("[CPOINT]: Not found")
	ErrCpointAlreadyRedeemed    = errors.New("[CPOINT]: Already redeemed")
	ErrCpointExpired            = errors.New("[CPOINT]: Expired")
	ErrCpointNotOwnedByCustomer = errors.New("[CPOINT]: Code does not belong to this customer")
)

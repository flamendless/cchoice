package errs

import "errors"

var (
	ErrCpointNotFound               = errors.New("[CPOINT]: Not found")
	ErrCpointAlreadyRedeemed        = errors.New("[CPOINT]: Already redeemed")
	ErrCpointExpired                = errors.New("[CPOINT]: Expired")
	ErrCpointNotOwnedByCustomer     = errors.New("[CPOINT]: Code does not belong to this customer")
	ErrCpointInvalidTokenFormat     = errors.New("[CPOINT]: Invalid token format")
	ErrCpointInvalidPayloadEncoding = errors.New("[CPOINT]: Invalid payload encoding")
	ErrCpointInvalidSignature       = errors.New("[CPOINT]: Invalid signature")
	ErrCpointInvalidPayload         = errors.New("[CPOINT]: Invalid payload")
	ErrCpointMissingRequiredFields  = errors.New("[CPOINT]: Missing required fields")
	ErrCpointTokenExpired           = errors.New("[CPOINT]: Token expired")
)

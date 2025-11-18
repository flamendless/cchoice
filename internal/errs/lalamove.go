package errs

import "errors"

var (
	ErrLalamoveServiceInit    = errors.New("[LALAMOVE]: Service must be configured")
	ErrLalamoveAPIKeyRequired = errors.New("[LALAMOVE]: API key required")
	ErrLalamoveSignRequest    = errors.New("[LALAMOVE]: Error in signing request")
	ErrLalamoveHTTPRequest    = errors.New("[LALAMOVE]: HTTP request failed")
	ErrLalamoveAPIResponse    = errors.New("[LALAMOVE]: API response error")
	ErrLalamoveJSONMarshal    = errors.New("[LALAMOVE]: JSON marshaling failed")
	ErrLalamoveJSONUnmarshal  = errors.New("[LALAMOVE]: JSON unmarshaling failed")
	ErrLalamoveQuotation      = errors.New("[LALAMOVE]: Quotation request failed")
	ErrLalamoveOrderCreate    = errors.New("[LALAMOVE]: Order creation failed")
	ErrLalamoveOrderStatus    = errors.New("[LALAMOVE]: Order status request failed")
	ErrLalamoveCapabilities   = errors.New("[LALAMOVE]: Capabilities request failed")
)

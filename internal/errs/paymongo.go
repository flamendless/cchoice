package errs

import "errors"

var (
	ErrPaymongoServiceInit             = errors.New("[PAYMONGO]: Service must be configured")
	ErrPaymongoAPIKeyRequired          = errors.New("[PAYMONGO]: API key required")
	ErrPaymongoAPIKeyInvalid           = errors.New("[PAYMONGO]: API key does not match environment")
	ErrPaymongoWebhookSignatureInvalid = errors.New("[PAYMONGO]: Webhook signature verification failed")
	ErrPaymongoWebhookPayloadInvalid   = errors.New("[PAYMONGO]: Webhook payload is invalid")
	ErrPaymongoWebhookSecretRequired   = errors.New("[PAYMONGO]: Webhook secret key is required")
)

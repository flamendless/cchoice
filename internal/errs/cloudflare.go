package errs

import "errors"

var (
	ErrCloudflareServiceInit  = errors.New("[CLOUDFLARE]: Service must be configured")
	ErrCloudflareAccountID    = errors.New("[CLOUDFLARE]: CLOUDFLARE_ACCOUNT_ID is required")
	ErrCloudflareAccountHash  = errors.New("[CLOUDFLARE]: CLOUDFLARE_ACCOUNT_HASH is required")
	ErrCloudflareAPIToken     = errors.New("[CLOUDFLARE]: CLOUDFLARE_IMAGES_API_TOKEN is required")
	ErrCloudflareUpload       = errors.New("[CLOUDFLARE]: Failed to upload image")
	ErrCloudflareDelete       = errors.New("[CLOUDFLARE]: Failed to delete image")
	ErrCloudflareGet          = errors.New("[CLOUDFLARE]: Failed to get image")
	ErrCloudflareAPI          = errors.New("[CLOUDFLARE]: API error")
	ErrCloudflareVerifyAccess = errors.New("[CLOUDFLARE]: Failed to verify access")
)

package constants

const (
	DefaultThumbnailSize = "96x96"
	CacheStore           = "cache store"
	Singleflight         = "singleflight"
	Gob                  = "gob"
	ViberURIPrefix       = "viber://chat?number="
	PHP                  = "PHP"
	PHMobilePrefix       = "+63"
	PrefixHTTPS          = "https://"

	// QueryParamAssetVersion carries the static asset fingerprint. Requests
	// that include it are served with an immutable cache policy.
	QueryParamAssetVersion = "v"
)

var CDNExcludePrefixKey = []string{
	"payments-",
	"brand_logos-",
}

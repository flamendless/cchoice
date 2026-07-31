package constants

const (
	// OriginImageCDN is the origin every remote image is served from. It is
	// preconnected in the document head so the TLS handshake is not paid for on
	// the first image request.
	OriginImageCDN = "https://imagedelivery.net"

	// PathImageCDNBase is the account-scoped prefix for delivery URLs.
	PathImageCDNBase = OriginImageCDN + "/YnES7emCTPeSEVA2N0dB_g"
)

const (
	EmptyImageFilename          = "empty_96x96.webp"
	PathEmptyImage              = "static/images/empty_96x96.webp"
	PathProductImages           = "static/images/product_images/"
	PathPaymentImages           = "static/images/payments/"
	PathEmailLogoCDN            = PathImageCDNBase + "/c89c033a-ebd3-4eed-0b79-09bfbcd16800/public"
	PathEmptyImageCDN           = PathImageCDNBase + "/empty_96x96/public"
	PathSVGLogoOnly             = PathImageCDNBase + "/ff4efd48-4b50-4763-2e18-d29f8afeab00/public"
	PathSVGLogoWithCompleteText = PathImageCDNBase + "/09edef59-3e15-4b00-4573-ece6e98d2800/public"
	PathStoreImageCDN           = PathImageCDNBase + "/store/public"
	PathCORSealImageCDN         = PathImageCDNBase + "/CORSeal/public"
	PathChangelogs              = "./CHANGELOGS.md"
	MaxImagePathLength          = 512
	PathLogoSmallLocal          = "cmd/web/static/svg/logo_only_small.png"
)

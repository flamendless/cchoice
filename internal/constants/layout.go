package constants

// Breakpoint used to switch between the mobile and desktop layouts. Keep in sync
// with tailwind's `lg` breakpoint.
const (
	MediaQueryDesktop = "(min-width: 1024px)"
	MediaQueryMobile  = "(max-width: 1023px)"
)

// Intrinsic dimensions of the fixed images rendered before their bytes arrive.
// Declaring them in the markup lets the browser reserve the right box and keeps
// the layout from shifting once the image decodes.
const (
	LogoWithTextWidth  = "719"
	LogoWithTextHeight = "238"
	LogoOnlyWidth      = "569"
	LogoOnlyHeight     = "768"

	StoreImageWidth  = "768"
	StoreImageHeight = "768"

	// BrandLogoWidth and BrandLogoHeight describe the fixed box brand logos are
	// drawn into rather than their intrinsic size, which differs per brand.
	BrandLogoWidth  = "256"
	BrandLogoHeight = "96"

	// ProductThumbnailSize is the rendered size of a product grid thumbnail at
	// the largest breakpoint.
	ProductThumbnailSize = "96"
)

// The `sizes` descriptors below tell the browser how wide each image renders so
// it can pick the smallest usable candidate out of the accompanying srcset.
const (
	// SizesSaleBannerImage matches the sale modal, capped at tailwind's `max-w-md`.
	SizesSaleBannerImage = "(max-width: 448px) 100vw, 448px"

	// SizesPromoBanner matches `w-full lg:w-1/2`.
	SizesPromoBanner = "(min-width: 1024px) 50vw, 100vw"

	// SizesStorePhoto matches the carousel, capped at tailwind's `max-w-lg`.
	SizesStorePhoto = "(min-width: 512px) 512px, 100vw"

	// SizesBrandLogo matches the fixed logo box.
	SizesBrandLogo = "256px"

	// SizesProductThumbnail matches `w-20 sm:w-24`.
	SizesProductThumbnail = "(min-width: 640px) 96px, 80px"
)

// Candidate widths requested from the image CDN for each responsive image. They
// cover 1x through roughly 3x of the rendered box.
var (
	WidthsPromoBanner      = []int{640, 960, 1366}
	WidthsStorePhoto       = []int{512, 768}
	WidthsBrandLogo        = []int{256, 512}
	WidthsProductThumbnail = []int{96, 192, 288}
)

// Widths used for the plain `src` of a responsive image, which browsers without
// srcset support fall back to.
const (
	WidthPromoBannerDefault      = 960
	WidthStorePhotoDefault       = 768
	WidthBrandLogoDefault        = 512
	WidthProductThumbnailDefault = 192
)

// Widths requested for the header logos, which render at a fixed height and are
// preloaded, so the preload and the img must ask for exactly the same URL.
const (
	WidthLogoWithText = 800
	WidthLogoOnly     = 256
)

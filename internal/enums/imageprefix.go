package enums

//go:generate go tool stringer -type=ImagePrefix -trimprefix=IMAGE_PREFIX_

type ImagePrefix int

const (
	IMAGE_PREFIX_UNDEFINED ImagePrefix = iota
	IMAGE_PREFIX_PRODUCT_IMAGE
	IMAGE_PREFIX_BRAND_IMAGE
	IMAGE_PREFIX_PROMO_IMAGE
)

func ParseImagePrefix(e string) ImagePrefix {
	switch e {
	case "PRODUCT_IMAGE":
		return IMAGE_PREFIX_PRODUCT_IMAGE
	case "BRAND_IMAGE":
		return IMAGE_PREFIX_BRAND_IMAGE
	case "PROMO_IMAGE":
		return IMAGE_PREFIX_PROMO_IMAGE
	default:
		return IMAGE_PREFIX_UNDEFINED
	}
}

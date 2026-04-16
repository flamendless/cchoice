package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=PromoType -trimprefix=PROMO_TYPE_

type PromoType int

const (
	PROMO_TYPE_UNDEFINED PromoType = iota
	PROMO_TYPE_BANNER_IMAGE
	PROMO_TYPE_BANNER_VIDEO
)

var AllPromoTypes = []PromoType{
	PROMO_TYPE_BANNER_IMAGE,
	PROMO_TYPE_BANNER_VIDEO,
}

func ParsePromoTypeToEnum(pt string) PromoType {
	switch strings.ToUpper(pt) {
	case PROMO_TYPE_BANNER_IMAGE.String():
		return PROMO_TYPE_BANNER_IMAGE
	case PROMO_TYPE_BANNER_VIDEO.String():
		return PROMO_TYPE_BANNER_VIDEO
	default:
		return PROMO_TYPE_UNDEFINED
	}
}

func MustParsePromoTypeToEnum(pt string) PromoType {
	res := ParsePromoTypeToEnum(pt)
	if res == PROMO_TYPE_UNDEFINED {
		panic(fmt.Sprintf("Unexpected PromoType. Got '%s'", pt))
	}
	return res
}

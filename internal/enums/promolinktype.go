package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=PromoLinkType -trimprefix=PROMO_LINK_TYPE_

type PromoLinkType int

const (
	PROMO_LINK_TYPE_UNDEFINED PromoLinkType = iota
	PROMO_LINK_TYPE_NONE
	PROMO_LINK_TYPE_TRACKED
	PROMO_LINK_TYPE_CUSTOM
)

var AllPromoLinkTypes = []PromoLinkType{
	PROMO_LINK_TYPE_NONE,
	PROMO_LINK_TYPE_TRACKED,
	PROMO_LINK_TYPE_CUSTOM,
}

func ParsePromoLinkTypeToEnum(s string) PromoLinkType {
	switch strings.ToUpper(s) {
	case PROMO_LINK_TYPE_NONE.String():
		return PROMO_LINK_TYPE_NONE
	case PROMO_LINK_TYPE_TRACKED.String():
		return PROMO_LINK_TYPE_TRACKED
	case PROMO_LINK_TYPE_CUSTOM.String():
		return PROMO_LINK_TYPE_CUSTOM
	default:
		return PROMO_LINK_TYPE_UNDEFINED
	}
}

func MustParsePromoLinkTypeToEnum(s string) PromoLinkType {
	res := ParsePromoLinkTypeToEnum(s)
	if res == PROMO_LINK_TYPE_UNDEFINED {
		panic(fmt.Sprintf("Unexpected PromoLinkType. Got '%s'", s))
	}
	return res
}

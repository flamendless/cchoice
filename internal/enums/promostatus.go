package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=PromoStatus -trimprefix=PROMO_STATUS_

type PromoStatus int

const (
	PROMO_STATUS_UNDEFINED PromoStatus = iota
	PROMO_STATUS_DRAFT
	PROMO_STATUS_PUBLISHED
	PROMO_STATUS_DELETED
)

var AllPromoStatuses = []PromoStatus{
	PROMO_STATUS_DRAFT,
	PROMO_STATUS_PUBLISHED,
	PROMO_STATUS_DELETED,
}

func ParsePromoStatusToEnum(ps string) PromoStatus {
	switch strings.ToUpper(ps) {
	case PROMO_STATUS_DRAFT.String():
		return PROMO_STATUS_DRAFT
	case PROMO_STATUS_PUBLISHED.String():
		return PROMO_STATUS_PUBLISHED
	case PROMO_STATUS_DELETED.String():
		return PROMO_STATUS_DELETED
	default:
		return PROMO_STATUS_UNDEFINED
	}
}

func MustParsePromoStatusToEnum(ps string) PromoStatus {
	res := ParsePromoStatusToEnum(ps)
	if res == PROMO_STATUS_UNDEFINED {
		panic(fmt.Sprintf("Unexpected PromoStatus. Got '%s'", ps))
	}
	return res
}

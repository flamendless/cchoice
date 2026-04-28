package enums

import "strings"

//go:generate go tool stringer -type=BrandStatus -trimprefix=BRAND_STATUS_

type BrandStatus int

const (
	BRAND_STATUS_UNDEFINED BrandStatus = iota
	BRAND_STATUS_DELETED
	BRAND_STATUS_DRAFT
	BRAND_STATUS_ACTIVE
)

var AllBrandStatuses = []BrandStatus{
	BRAND_STATUS_DELETED,
	BRAND_STATUS_DRAFT,
	BRAND_STATUS_ACTIVE,
}

func ParseBrandStatusToEnum(e string) BrandStatus {
	switch strings.ToUpper(e) {
	case BRAND_STATUS_DELETED.String():
		return BRAND_STATUS_DELETED
	case BRAND_STATUS_DRAFT.String():
		return BRAND_STATUS_DRAFT
	case BRAND_STATUS_ACTIVE.String():
		return BRAND_STATUS_ACTIVE
	default:
		return BRAND_STATUS_UNDEFINED
	}
}

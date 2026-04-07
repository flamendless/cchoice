package enums

//go:generate go tool stringer -type=ProductStatus -trimprefix=PRODUCT_STATUS_

type ProductStatus int

const (
	PRODUCT_STATUS_UNDEFINED ProductStatus = iota
	PRODUCT_STATUS_ACTIVE
	PRODUCT_STATUS_DELETED
	PRODUCT_STATUS_DRAFT
)

func ParseProductStatusToEnum(e string) ProductStatus {
	switch e {
	case PRODUCT_STATUS_ACTIVE.String():
		return PRODUCT_STATUS_ACTIVE
	case PRODUCT_STATUS_DELETED.String():
		return PRODUCT_STATUS_DELETED
	case PRODUCT_STATUS_DRAFT.String():
		return PRODUCT_STATUS_DRAFT
	default:
		return PRODUCT_STATUS_UNDEFINED
	}
}

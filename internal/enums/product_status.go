package enums

type ProductStatus int

const (
	PRODUCT_STATUS_UNDEFINED ProductStatus = iota
	PRODUCT_STATUS_ACTIVE
	PRODUCT_STATUS_DELETED
)

func (t ProductStatus) String() string {
	switch t {
	case PRODUCT_STATUS_ACTIVE:
		return "ACTIVE"
	case PRODUCT_STATUS_DELETED:
		return "DELETED"
	default:
		return "UNDEFINED"
	}
}

func ParseProductStatusEnum(e string) ProductStatus {
	switch e {
	case "ACTIVE":
		return PRODUCT_STATUS_ACTIVE
	case "DELETED":
		return PRODUCT_STATUS_DELETED
	default:
		return PRODUCT_STATUS_UNDEFINED
	}
}

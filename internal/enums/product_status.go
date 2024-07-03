package enums

import (
	pb "cchoice/proto"
)

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

func ParseProductStatusEnumPB(e string) pb.ProductStatus_ProductStatus {
	switch e {
	case "ACTIVE":
		return pb.ProductStatus_ACTIVE
	case "DELETED":
		return pb.ProductStatus_DELETED
	default:
		return pb.ProductStatus_UNDEFINED
	}
}

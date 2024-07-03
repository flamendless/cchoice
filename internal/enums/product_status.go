package enums

import (
	pb "cchoice/proto"
	"fmt"
)

type ProductStatus int

const (
	PRODUCT_STATUS_UNDEFINED ProductStatus = iota
	PRODUCT_STATUS_ACTIVE
	PRODUCT_STATUS_DELETED
)

func (t ProductStatus) String() string {
	switch t {
	case PRODUCT_STATUS_UNDEFINED:
		return "UNDEFINED"
	case PRODUCT_STATUS_ACTIVE:
		return "ACTIVE"
	case PRODUCT_STATUS_DELETED:
		return "DELETED"
	default:
		panic("unknown enum")
	}
}

func ParseProductStatusEnum(e string) ProductStatus {
	switch e {
	case "UNDEFINED":
		return PRODUCT_STATUS_UNDEFINED
	case "ACTIVE":
		return PRODUCT_STATUS_ACTIVE
	case "DELETED":
		return PRODUCT_STATUS_DELETED
	default:
		panic(fmt.Sprintf("Can't convert '%s' to ProductStatus enum", e))
	}
}

func ParseProductStatusEnumPB(e string) pb.ProductStatus_ProductStatus {
	switch e {
	case "ACTIVE":
		return pb.ProductStatus_ACTIVE
	case "DELETED":
		return pb.ProductStatus_DELETED
	case "UNDEFINED":
		return pb.ProductStatus_UNDEFINED
	default:
		panic(fmt.Sprintf("Can't convert '%s' to pb.ProductStatus enum", e))
	}
}

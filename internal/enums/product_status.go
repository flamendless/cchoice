package enums

import (
	pb "cchoice/proto"
	"fmt"
)

type ProductStatus int

const (
	UndefinedProductStatus ProductStatus = iota
	Active
	Deleted
)

func (t ProductStatus) String() string {
	switch t {
	case UndefinedProductStatus:
		return "UNDEFINED"
	case Active:
		return "ACTIVE"
	case Deleted:
		return "DELETED"
	default:
		panic("unknown enum")
	}
}

func ParseProductStatusEnum(e string) ProductStatus {
	switch e {
	case "UNDEFINED":
		return UndefinedProductStatus
	case "ACTIVE":
		return Active
	case "DELETED":
		return Deleted
	default:
		panic(fmt.Sprintf("Can't convert '%s' to ProductStatus enum", e))
	}
}

func ParseProductStatusEnumPB(e string) pb.ProductStatus {
	switch e {
	case "UNDEFINED":
		return pb.ProductStatus_PRODUCT_STATUS_UNDEFINED
	case "ACTIVE":
		return pb.ProductStatus_ACTIVE
	case "DELETED":
		return pb.ProductStatus_DELETED
	default:
		panic(fmt.Sprintf("Can't convert '%s' to pb.ProductStatus enum", e))
	}
}

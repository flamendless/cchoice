package enums

import (
	pb "cchoice/proto"
	"fmt"
)

type ProductStatus int

const (
	Undefined ProductStatus = iota
	Active
	Deleted
)

func (t ProductStatus) String() string {
	switch t {
	case Undefined:
		return "undefined"
	case Active:
		return "active"
	case Deleted:
		return "deleted"
	default:
		panic("unknown enum")
	}
}

func ParseProductStatusEnum(e string) ProductStatus {
	switch e {
	case "undefined":
		return Undefined
	case "active":
		return Active
	case "deleted":
		return Deleted
	default:
		panic(fmt.Sprintf("Can't convert '%s' to ProductStatus enum", e))
	}
}

func ParseProductStatusEnumPB(e string) pb.ProductStatus {
	switch e {
	case "undefined":
		return pb.ProductStatus_undefined
	case "active":
		return pb.ProductStatus_active
	case "deleted":
		return pb.ProductStatus_deleted
	default:
		panic(fmt.Sprintf("Can't convert '%s' to pb.ProductStatus enum", e))
	}
}

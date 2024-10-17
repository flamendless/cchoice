package enums

import (
	pb "cchoice/proto"
	"testing"

	"github.com/stretchr/testify/require"
)

var tblProductStatus = map[ProductStatus]string{
	PRODUCT_STATUS_UNDEFINED: "UNDEFINED",
	PRODUCT_STATUS_ACTIVE:    "ACTIVE",
	PRODUCT_STATUS_DELETED:   "DELETED",
}

func TestProductStatusToString(t *testing.T) {
	for productstatus, val := range tblProductStatus {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, val, productstatus.String())
		})
	}
}

func TestParseProductStatusEnum(t *testing.T) {
	for productstatus, val := range tblProductStatus {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, productstatus, ParseProductStatusEnum(val))
		})
	}
}

func TestParseProductStatusEnumPB(t *testing.T) {
	tbl := map[string]pb.ProductStatus_ProductStatus{
		"UNDEFINED": pb.ProductStatus_UNDEFINED,
		"ACTIVE":    pb.ProductStatus_ACTIVE,
		"DELETED":   pb.ProductStatus_DELETED,
	}
	for val, productstatus := range tbl {
		t.Run(val, func(t *testing.T) {
			enum := StringToPBEnum(val, pb.ProductStatus_ProductStatus_value, pb.ProductStatus_UNDEFINED)
			require.Equal(t, productstatus, enum)
		})
	}
}

func BenchmarkProductStatusToString(b *testing.B) {
	for sortfield := range tblProductStatus {
		for i := 0; i < b.N; i++ {
			_ = sortfield.String()
		}
	}
}

func BenchmarkParseProductStatusEnum(b *testing.B) {
	for _, val := range tblProductStatus {
		for i := 0; i < b.N; i++ {
			_ = ParseProductStatusEnum(val)
		}
	}
}

func BenchmarkParseProductStatusEnumPB(b *testing.B) {
	tbl := map[string]pb.ProductStatus_ProductStatus{
		"UNDEFINED": pb.ProductStatus_UNDEFINED,
		"ACTIVE":    pb.ProductStatus_ACTIVE,
		"DELETED":   pb.ProductStatus_DELETED,
	}
	for val := range tbl {
		for i := 0; i < b.N; i++ {
			_ = StringToPBEnum(val, pb.ProductStatus_ProductStatus_value, pb.ProductStatus_UNDEFINED)
		}
	}
}

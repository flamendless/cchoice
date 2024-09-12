package enums

import (
	pb "cchoice/proto"
	"testing"
)

var tblProductStatus = map[ProductStatus]string{
	PRODUCT_STATUS_UNDEFINED: "UNDEFINED",
	PRODUCT_STATUS_ACTIVE:    "ACTIVE",
	PRODUCT_STATUS_DELETED:   "DELETED",
}

func TestProductStatusToString(t *testing.T) {
	for productstatus, val := range tblProductStatus {
		if productstatus.String() != val {
			t.Fatalf("Mismatch: %s = %s", productstatus.String(), val)
		}
	}
}

func TestParseProductStatusEnum(t *testing.T) {
	for productstatus, val := range tblProductStatus {
		parsed := ParseProductStatusEnum(val)
		if parsed != productstatus {
			t.Fatalf("Mismatch: %s = %s", val, productstatus)
		}
	}
}

func TestParseProductStatusEnumPB(t *testing.T) {
	tbl := map[string]pb.ProductStatus_ProductStatus{
		"UNDEFINED": pb.ProductStatus_UNDEFINED,
		"ACTIVE":    pb.ProductStatus_ACTIVE,
		"DELETED":   pb.ProductStatus_DELETED,
	}
	for val, productstatus := range tbl {
		enum := StringToPBEnum(val, pb.ProductStatus_ProductStatus_value, pb.ProductStatus_UNDEFINED)
		if enum != productstatus {
			t.Fatalf("Mismatch: %s = %s", enum, productstatus)
		}
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

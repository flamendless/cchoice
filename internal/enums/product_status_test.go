package enums

import (
	pb "cchoice/proto"
	"testing"
)

func TestProductStatusToString(t *testing.T) {
	undef := PRODUCT_STATUS_UNDEFINED
	active := PRODUCT_STATUS_ACTIVE
	deleted := PRODUCT_STATUS_DELETED

	if undef.String() != "UNDEFINED" {
		t.Fatalf("Mismatch: %s = %s", undef.String(), "UNDEFINED")
	}

	if active.String() != "ACTIVE" {
		t.Fatalf("Mismatch: %s = %s", active.String(), "ACTIVE")
	}

	if deleted.String() != "DELETED" {
		t.Fatalf("Mismatch: %s = %s", deleted.String(), "DELETED")
	}
}

func TestParseProductStatusEnum(t *testing.T) {
	undef := ParseProductStatusEnum("UNDEFINED")
	asc := ParseProductStatusEnum("ACTIVE")
	desc := ParseProductStatusEnum("DELETED")

	if undef != PRODUCT_STATUS_UNDEFINED {
		t.Fatalf("Mismatch: %s = %s", undef, PRODUCT_STATUS_UNDEFINED)
	}
	if asc != PRODUCT_STATUS_ACTIVE {
		t.Fatalf("Mismatch: %s = %s", asc, PRODUCT_STATUS_ACTIVE)
	}
	if desc != PRODUCT_STATUS_DELETED {
		t.Fatalf("Mismatch: %s = %s", desc, PRODUCT_STATUS_DELETED)
	}
}

func TestParseProductStatusEnumPB(t *testing.T) {
	undef := StringToPBEnum(
		"UNDEFINED",
		pb.ProductStatus_ProductStatus_value,
		pb.ProductStatus_UNDEFINED,
	)
	active := StringToPBEnum(
		"ACTIVE",
		pb.ProductStatus_ProductStatus_value,
		pb.ProductStatus_UNDEFINED,
	)
	deleted := StringToPBEnum(
		"DELETED",
		pb.ProductStatus_ProductStatus_value,
		pb.ProductStatus_UNDEFINED,
	)

	if undef != pb.ProductStatus_UNDEFINED {
		t.Fatalf("Mismatch: %s = %s", undef, pb.ProductStatus_UNDEFINED)
	}
	if active != pb.ProductStatus_ACTIVE {
		t.Fatalf("Mismatch: %s = %s", active, pb.ProductStatus_ACTIVE)
	}
	if deleted != pb.ProductStatus_DELETED {
		t.Fatalf("Mismatch: %s = %s", deleted, pb.ProductStatus_DELETED)
	}
}

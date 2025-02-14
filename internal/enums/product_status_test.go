package enums

import (
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

func BenchmarkProductStatusToString(b *testing.B) {
	for sortfield := range tblProductStatus {
		for b.Loop() {
			_ = sortfield.String()
		}
	}
}

func BenchmarkParseProductStatusEnum(b *testing.B) {
	for _, val := range tblProductStatus {
		for b.Loop() {
			_ = ParseProductStatusEnum(val)
		}
	}
}

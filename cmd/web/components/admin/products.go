package components

import "cchoice/internal/enums"

func productListRowClass(status enums.ProductStatus) string {
	switch status {
	case enums.PRODUCT_STATUS_DELETED:
		return "bg-red-50"
	case enums.PRODUCT_STATUS_DRAFT:
		return "bg-yellow-50"
	case enums.PRODUCT_STATUS_ACTIVE:
		return "bg-green-50"
	default:
		return ""
	}
}

func productListTimestamp(updatedAt, createdAt string) string {
	if updatedAt != "" && updatedAt != createdAt {
		return updatedAt
	}
	return createdAt
}

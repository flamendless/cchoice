// Code generated by "stringer -type=ProductStatus -trimprefix=PRODUCT_STATUS_"; DO NOT EDIT.

package enums

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PRODUCT_STATUS_UNDEFINED-0]
	_ = x[PRODUCT_STATUS_ACTIVE-1]
	_ = x[PRODUCT_STATUS_DELETED-2]
}

const _ProductStatus_name = "UNDEFINEDACTIVEDELETED"

var _ProductStatus_index = [...]uint8{0, 9, 15, 22}

func (i ProductStatus) String() string {
	if i < 0 || i >= ProductStatus(len(_ProductStatus_index)-1) {
		return "ProductStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ProductStatus_name[_ProductStatus_index[i]:_ProductStatus_index[i+1]]
}
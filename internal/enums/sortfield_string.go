// Code generated by "stringer -type=SortField -trimprefix=SORT_FIELD_"; DO NOT EDIT.

package enums

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[SORT_FIELD_UNDEFINED-0]
	_ = x[SORT_FIELD_NAME-1]
	_ = x[SORT_FIELD_CREATED_AT-2]
}

const _SortField_name = "UNDEFINEDNAMECREATED_AT"

var _SortField_index = [...]uint8{0, 9, 13, 23}

func (i SortField) String() string {
	if i < 0 || i >= SortField(len(_SortField_index)-1) {
		return "SortField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _SortField_name[_SortField_index[i]:_SortField_index[i+1]]
}
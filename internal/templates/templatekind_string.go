// Code generated by "stringer -type=TemplateKind -trimprefix=TEMPLATE_"; DO NOT EDIT.

package templates

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TEMPLATE_UNDEFINED-0]
	_ = x[TEMPLATE_SAMPLE-1]
	_ = x[TEMPLATE_DELTAPLUS-2]
	_ = x[TEMPLATE_BOSCH-3]
	_ = x[TEMPLATE_SPARTAN-4]
	_ = x[TEMPLATE_SHINSETSU-5]
	_ = x[TEMPLATE_REDMAX-6]
	_ = x[TEMPLATE_BRADFORD-7]
	_ = x[TEMPLATE_KOBEWEL-8]
}

const _TemplateKind_name = "UNDEFINEDSAMPLEDELTAPLUSBOSCHSPARTANSHINSETSUREDMAXBRADFORDKOBEWEL"

var _TemplateKind_index = [...]uint8{0, 9, 15, 24, 29, 36, 45, 51, 59, 66}

func (i TemplateKind) String() string {
	if i < 0 || i >= TemplateKind(len(_TemplateKind_index)-1) {
		return "TemplateKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TemplateKind_name[_TemplateKind_index[i]:_TemplateKind_index[i+1]]
}

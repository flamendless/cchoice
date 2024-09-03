package templates

import (
	"testing"
)

var tblTemplateKind = map[TemplateKind]string{
	TEMPLATE_UNDEFINED: "UNDEFINED",
	TEMPLATE_SAMPLE:    "SAMPLE",
	TEMPLATE_DELTAPLUS: "DELTAPLUS",
	TEMPLATE_BOSCH:     "BOSCH",
	TEMPLATE_SPARTAN:   "SPARTAN",
	TEMPLATE_SHINSETSU: "SHINSETSU",
	TEMPLATE_REDMAX:    "REDMAX",
	TEMPLATE_BRADFORD:  "BRADFORD",
	TEMPLATE_KOBEWEL:   "KOBEWEL",
}

func TestTemplateEnumToString(t *testing.T) {
	for tpl, val := range tblTemplateKind {
		if tpl.String() != val {
			t.Fatalf("Mismatch: %s = %s", tpl.String(), val)
		}
	}
}

func TestParseTemplateEnum(t *testing.T) {
	for tpl, val := range tblTemplateKind {
		parsed := ParseTemplateEnum(val)
		if parsed != tpl {
			t.Fatalf("Mismatch: %s = %s", val, tpl)
		}
	}
}

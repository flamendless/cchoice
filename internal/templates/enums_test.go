package templates

import (
	"testing"

	"github.com/stretchr/testify/require"
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
		t.Run(val, func(t *testing.T) {
			require.Equal(t, val, tpl.String())
		})
	}
}

func TestParseTemplateEnum(t *testing.T) {
	for tpl, val := range tblTemplateKind {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, tpl, ParseTemplateEnum(val))
		})
	}
}

func BenchmarkTemplateToString(b *testing.B) {
	for tpl := range tblTemplateKind {
		for i := 0; i < b.N; i++ {
			_ = tpl.String()
		}
	}
}

func BenchmarkParseTemplateEnum(b *testing.B) {
	for _, val := range tblTemplateKind {
		for i := 0; i < b.N; i++ {
			_ = ParseTemplateEnum(val)
		}
	}
}

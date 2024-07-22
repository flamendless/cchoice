package templates

import "fmt"

type TemplateKind int

const (
	TEMPLATE_UNDEFINED TemplateKind = iota
	TEMPLATE_SAMPLE
	TEMPLATE_DELTAPLUS
	TEMPLATE_BOSCH
)

func (t TemplateKind) String() string {
	switch t {
	case TEMPLATE_UNDEFINED:
		return "undefined"
	case TEMPLATE_SAMPLE:
		return "sample"
	case TEMPLATE_DELTAPLUS:
		return "delta_plus"
	case TEMPLATE_BOSCH:
		return "bosch"
	default:
		panic("unknown enum")
	}
}

func ParseTemplateEnum(e string) TemplateKind {
	switch e {
	case "undefined":
		return TEMPLATE_UNDEFINED
	case "sample":
		return TEMPLATE_SAMPLE
	case "delta_plus":
		return TEMPLATE_DELTAPLUS
	case "bosch":
		return TEMPLATE_BOSCH
	default:
		panic(fmt.Sprintf("Can't convert '%s' to TemplateKind enum", e))
	}
}

func TemplateToBrand(tpl string) string {
	switch tpl {
	case "sample":
		return "sample"
	case "delta_plus":
		return "DeltaPlus"
	case "bosch":
		return "Bosch"
	default:
		panic(fmt.Sprintf("Can't convert template '%s' to brand string", tpl))
	}
}

package templates

import "fmt"

type TemplateKind int

const (
	TEMPLATE_UNDEFINED TemplateKind = iota
	TEMPLATE_SAMPLE
	TEMPLATE_DELTAPLUS
	TEMPLATE_BOSCH
	TEMPLATE_SPARTA
	TEMPLATE_SHINSETSU
	TEMPLATE_REDMAX
	TEMPLATE_BRADFORD
	TEMPLATE_KOBEWEL
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
	case TEMPLATE_SPARTA:
		return "sparta"
	case TEMPLATE_SHINSETSU:
		return "shinsetsu"
	case TEMPLATE_REDMAX:
		return "redmax"
	case TEMPLATE_BRADFORD:
		return "bradford"
	case TEMPLATE_KOBEWEL:
		return "kobewel"
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
	case "sparta":
		return TEMPLATE_SPARTA
	case "shinsetsu":
		return TEMPLATE_SHINSETSU
	case "redmax":
		return TEMPLATE_REDMAX
	case "bradford":
		return TEMPLATE_BRADFORD
	case "kobewel":
		return TEMPLATE_KOBEWEL
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
	case "sparta":
		return "Sparta"
	case "shinsetsu":
		return "Shinsetsu"
	case "redmax":
		return "RedMax"
	case "bradford":
		return "Brandford"
	case "kobewel":
		return "Kobewel"
	default:
		panic(fmt.Sprintf("Can't convert template '%s' to brand string", tpl))
	}
}

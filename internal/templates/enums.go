package templates

import "fmt"

type TemplateKind int

const (
	Undefined TemplateKind = iota
	Sample
	DeltaPlus
	Bosch
)

func (t TemplateKind) String() string {
	switch t {
	case Undefined:
		return "undefined"
	case Sample:
		return "sample"
	case DeltaPlus:
		return "delta_plus"
	case Bosch:
		return "bosch"
	default:
		panic("unknown enum")
	}
}

func ParseTemplateEnum(e string) TemplateKind {
	switch e {
	case "undefined":
		return Undefined
	case "sample":
		return Sample
	case "delta_plus":
		return DeltaPlus
	case "bosch":
		return Bosch
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
		return ""
	}
}

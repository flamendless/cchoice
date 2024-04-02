package templates

import "fmt"

type TemplateKind int

const (
	Undefined TemplateKind = iota
	Sample
	DeltaPlus
)

func (t TemplateKind) String() string {
	switch t {
	case Undefined:
		return "undefined"
	case Sample:
		return "sample"
	case DeltaPlus:
		return "delta_plus"
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
	default:
		panic(fmt.Sprintf("Can't convert '%s' to TemplateKind enum", e))
	}
}

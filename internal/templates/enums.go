package templates

import "fmt"

type TemplateKind int

const (
	Undefined     TemplateKind = iota
	Sample
)

func (t TemplateKind) String() string {
	switch t {
	case Undefined:
		return "undefined"
	case Sample:
		return "sample"
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
	default:
		panic(fmt.Sprintf("Can't convert '%s' to TemplateKind enum", e))
	}
}

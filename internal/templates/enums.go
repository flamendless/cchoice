package templates

//go:generate stringer -type=TemplateKind -linecomment

import "fmt"

type TemplateKind int

const (
	TEMPLATE_UNDEFINED TemplateKind = iota //undefined
	TEMPLATE_SAMPLE                        //sample
	TEMPLATE_DELTAPLUS                     //delta_plus
	TEMPLATE_BOSCH                         //bosch
	TEMPLATE_SPARTA                        //sparta
	TEMPLATE_SHINSETSU                     //shinsetsu
	TEMPLATE_REDMAX                        //redmax
	TEMPLATE_BRADFORD                      //bradford
	TEMPLATE_KOBEWEL                       //kobewel
)

func ParseTemplateEnum(e string) TemplateKind {
	switch e {
	case TEMPLATE_UNDEFINED.String():
		return TEMPLATE_UNDEFINED
	case TEMPLATE_SAMPLE.String():
		return TEMPLATE_SAMPLE
	case TEMPLATE_DELTAPLUS.String():
		return TEMPLATE_DELTAPLUS
	case TEMPLATE_BOSCH.String():
		return TEMPLATE_BOSCH
	case TEMPLATE_SPARTA.String():
		return TEMPLATE_SPARTA
	case TEMPLATE_SHINSETSU.String():
		return TEMPLATE_SHINSETSU
	case TEMPLATE_REDMAX.String():
		return TEMPLATE_REDMAX
	case TEMPLATE_BRADFORD.String():
		return TEMPLATE_BRADFORD
	case TEMPLATE_KOBEWEL.String():
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

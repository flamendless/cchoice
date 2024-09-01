package templates

//go:generate stringer -type=TemplateKind -trimprefix=TEMPLATE_

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

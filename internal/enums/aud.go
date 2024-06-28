package enums

import "fmt"

type AudKind int

const (
	Undefined AudKind = iota
	AudAPI
	AudSystem
)

func (a AudKind) String() string {
	switch a {
	case Undefined:
		return "undefined"
	case AudAPI:
		return "API"
	case AudSystem:
		return "system"
	default:
		panic("unknown enum")
	}
}

func ParseAudEnum(e string) AudKind {
	switch e {
	case "undefined":
		return Undefined
	case "API":
		return AudAPI
	case "system":
		return AudSystem
	default:
		panic(fmt.Sprintf("Can't convert '%s' to AudKind enum", e))
	}
}

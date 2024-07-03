package enums

type AudKind int

const (
	AUD_UNDEFINED AudKind = iota
	AUD_API
	AUD_SYSTEM
)

func (a AudKind) String() string {
	switch a {
	case AUD_UNDEFINED:
		return "undefined"
	case AUD_API:
		return "API"
	case AUD_SYSTEM:
		return "system"
	default:
		return "undefined"
	}
}

func ParseAudEnum(e string) AudKind {
	switch e {
	case "undefined":
		return AUD_UNDEFINED
	case "API":
		return AUD_API
	case "system":
		return AUD_SYSTEM
	default:
		return AUD_UNDEFINED
	}
}

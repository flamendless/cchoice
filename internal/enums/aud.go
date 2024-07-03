package enums

type AudKind int

const (
	AUD_UNDEFINED AudKind = iota
	AUD_API
	AUD_SYSTEM
)

func (a AudKind) String() string {
	switch a {
	case AUD_API:
		return "API"
	case AUD_SYSTEM:
		return "SYSTEM"
	default:
		return "UNDEFINED"
	}
}

func ParseAudEnum(e string) AudKind {
	switch e {
	case "API":
		return AUD_API
	case "SYSTEM":
		return AUD_SYSTEM
	default:
		return AUD_UNDEFINED
	}
}

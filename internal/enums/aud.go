package enums

type AudKind int

const (
	AudUndefined AudKind = iota
	AudAPI
	AudSystem
)

func (a AudKind) String() string {
	switch a {
	case AudUndefined:
		return "undefined"
	case AudAPI:
		return "API"
	case AudSystem:
		return "system"
	default:
		return "undefined"
	}
}

func ParseAudEnum(e string) AudKind {
	switch e {
	case "undefined":
		return AudUndefined
	case "API":
		return AudAPI
	case "system":
		return AudSystem
	default:
		return AudUndefined
	}
}

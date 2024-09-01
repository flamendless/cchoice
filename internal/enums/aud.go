package enums

//go:generate stringer -type=AudKind -trimprefix=AUD_

type AudKind int

const (
	AUD_UNDEFINED AudKind = iota
	AUD_API
	AUD_SYSTEM
)

func ParseAudEnum(e string) AudKind {
	switch e {
	case AUD_API.String():
		return AUD_API
	case AUD_SYSTEM.String():
		return AUD_SYSTEM
	default:
		return AUD_UNDEFINED
	}
}

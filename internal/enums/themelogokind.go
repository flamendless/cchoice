package enums

// ThemeLogoKind identifies which of the two theme logo variants an upload
// is for. Its String() form is used verbatim as part of the Cloudflare
// object key, so it intentionally does not go through the stringer tool
// (which would uppercase and underscore the value).
type ThemeLogoKind int

const (
	THEME_LOGO_KIND_UNDEFINED ThemeLogoKind = iota
	THEME_LOGO_KIND_LOGO
	THEME_LOGO_KIND_LOGO_WITH_TEXT
)

func (k ThemeLogoKind) String() string {
	switch k {
	case THEME_LOGO_KIND_LOGO:
		return "logo"
	case THEME_LOGO_KIND_LOGO_WITH_TEXT:
		return "logowithtext"
	default:
		return "undefined"
	}
}

func ParseThemeLogoKindToEnum(k string) ThemeLogoKind {
	switch k {
	case THEME_LOGO_KIND_LOGO.String():
		return THEME_LOGO_KIND_LOGO
	case THEME_LOGO_KIND_LOGO_WITH_TEXT.String():
		return THEME_LOGO_KIND_LOGO_WITH_TEXT
	default:
		return THEME_LOGO_KIND_UNDEFINED
	}
}

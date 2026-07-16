package constants

var ThemeColorKeys = []string{
	"primary",
	"primary-dark",
	"primary-emphasis",
	"primary-hover",
	"primary-muted",
	"surface",
	"accent",
}

const (
	ThemeConfigKeyLogoURL         = "logo_url"
	ThemeConfigKeyLogoWithTextURL = "logo_with_text_url"
)

// MaxSizeThemeLogoUpload mirrors MaxSizeImageUpload; kept separate so theme
// logo limits can be tuned independently of other image uploads.
const MaxSizeThemeLogoUpload int64 = MaxSizeImageUpload

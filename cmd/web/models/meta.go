package models

import "cchoice/internal/utils"

type ProductsMeta struct {
	Title           string
	Content         string
	CanonicalURL    string
	OGImage         string
	OGType          string
	Robots          string
	Keywords        string
	StructuredData  string
	PriceAmount     string
	PriceCurrency   string
	TwitterCard     string
}

type SiteSEO struct {
	Title          string
	Description    string
	CanonicalURL   string
	Robots         string
	Keywords       string
	OGTitle        string
	OGDescription  string
	OGType         string
	OGURL          string
	OGImage        string
	TwitterCard    string
}

func DefaultSiteSEO() SiteSEO {
	canonical := utils.SiteURL("/")
	return SiteSEO{
		Title:         "C-Choice Construction Supply",
		Description:   "Your Partner in Progress. Quality construction tools and materials from trusted brands in the Philippines.",
		CanonicalURL:  canonical,
		Robots:        "index, follow",
		Keywords:      "c-choice, construction, power tools, philippines",
		OGTitle:       "C-Choice Construction Supplies",
		OGDescription: "Your Partner in Progress",
		OGType:        "website",
		OGURL:         canonical,
		OGImage:       "https://imagedelivery.net/YnES7emCTPeSEVA2N0dB_g/favicons-192x192/public",
		TwitterCard:   "summary",
	}
}

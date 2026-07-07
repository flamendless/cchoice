package models

import (
	"encoding/json"

	"cchoice/internal/seo"
	"cchoice/internal/utils"
)

type ProductsMeta struct {
	Title          string
	Content        string
	CanonicalURL   string
	OGImage        string
	OGType         string
	Robots         string
	Keywords       string
	StructuredData json.RawMessage
	PriceAmount    string
	PriceCurrency  string
	TwitterCard    string
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
	StructuredData json.RawMessage
}

func DefaultSiteSEO() SiteSEO {
	canonical := utils.SiteURL("/")
	ogImage := seo.DefaultOGImage
	return SiteSEO{
		Title:          seo.SiteTitle,
		Description:    seo.SiteDescription,
		CanonicalURL:   canonical,
		Robots:         "index, follow, max-image-preview:large",
		Keywords:       seo.SiteKeywords,
		OGTitle:        seo.SiteOGTitle,
		OGDescription:  seo.SiteOGDescription,
		OGType:         "website",
		OGURL:          canonical,
		OGImage:        ogImage,
		TwitterCard:    "summary_large_image",
		StructuredData: seo.BuildSiteStructuredData(canonical, ogImage),
	}
}

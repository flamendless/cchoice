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

func NotFoundSEO() SiteSEO {
	seoMeta := DefaultSiteSEO()
	seoMeta.Robots = "noindex, follow"
	seoMeta.CanonicalURL = ""
	seoMeta.OGURL = ""
	return seoMeta
}

func HomePageSEO(filters HomePageFilters) SiteSEO {
	seoMeta := DefaultSiteSEO()
	if filters.BrandID != "" {
		seoMeta.Robots = "noindex, follow"
	}
	return seoMeta
}

func CategoryPageSEO(categorySlug, subcategorySlug string) SiteSEO {
	meta := seo.GenerateCategoryMeta(categorySlug, subcategorySlug)
	return SiteSEO{
		Title:          meta.Title,
		Description:    meta.Description,
		CanonicalURL:   meta.CanonicalURL,
		Robots:         meta.Robots,
		Keywords:       meta.Keywords,
		OGTitle:        meta.Title,
		OGDescription:  meta.Description,
		OGType:         meta.OGType,
		OGURL:          meta.CanonicalURL,
		OGImage:        meta.OGImage,
		TwitterCard:    meta.TwitterCard,
		StructuredData: meta.StructuredData,
	}
}

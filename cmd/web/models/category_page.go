package models

type CategoryPageData struct {
	CategorySlug     string
	SubcategorySlug  string
	CategoryLabel    string
	SubcategoryLabel string
	Products         []CategorySectionProduct
	SEO              SiteSEO
	ThemeCSS         string
}

type CategorySitemapSlug struct {
	Category    string
	Subcategory string
}

package models

type BrandsListingPageData struct {
	Brands   []BrandListingCard
	SEO      SiteSEO
	ThemeCSS string
}

type BrandListingCard struct {
	Slug         string
	Name         string
	LogoURL      string
	ProductCount int64
	ComingSoon   bool
	PageURL      string
}

type BrandPageData struct {
	Slug              string
	Name              string
	LogoURL           string
	BestSelling       BrandPrioritySection
	HighestDiscount   BrandPrioritySection
	CategorySections  []BrandGroupedCategorySection
	SEO               SiteSEO
	ThemeCSS          string
}

type BrandPrioritySection struct {
	Title       string
	Products    []CategorySectionProduct
	TotalCount  int
	ShowMoreURL string
	HasMore     bool
}

type BrandGroupedCategorySection struct {
	Label          string
	ScrollTargetID string
	Subcategories  []BrandSubcategorySection
}

type BrandSubcategorySection struct {
	CategoryID  string
	Label       string
	ProductsURL string
}

type BrandSitemapSlug struct {
	Slug string
}

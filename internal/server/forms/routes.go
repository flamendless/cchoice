package forms

type ProductsImageQuery struct {
	Path      string `form:"path" validate:"required"`
	Thumbnail string `form:"thumbnail"`
	Size      string `form:"size"`
	Quality   string `form:"quality"`
}

type AssetFilenameQuery struct {
	Filename string `form:"filename" validate:"required"`
}

type HomeBrandQuery struct {
	BrandID string `form:"brand_id"`
}

type ChangelogsQuery struct {
	AppEnv string `form:"appenv"`
}

type SearchForm struct {
	Search       string `form:"search"`
	SearchMobile string `form:"search-mobile"`
}

type MetricsEventQuery struct {
	Event string `form:"event" validate:"required"`
	Value string `form:"value"`
}

package forms

type BrandPagePath struct {
	Brand string `param:"brand" validate:"required"`
}

type BrandCategoryProductsPath struct {
	Brand      string `param:"brand" validate:"required"`
	CategoryID string `param:"category_id" validate:"required"`
}

type BrandPrioritySectionPath struct {
	Brand   string `param:"brand" validate:"required"`
	Section string `param:"section" validate:"required"`
}

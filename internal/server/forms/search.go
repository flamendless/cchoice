package forms

type SearchPageQuery struct {
	Q string `form:"q" validate:"required,min_search"`
}

type SearchProductsQuery struct {
	Q    string `form:"q" validate:"required,min_search"`
	Page int    `form:"page"`
}

type SearchRelatedQuery struct {
	Q      string `form:"q" validate:"required,min_search"`
	Page   int    `form:"page"`
	Source string `form:"source" validate:"omitempty,oneof=related category brand"`
}

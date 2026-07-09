package forms

import "cchoice/internal/constants"

type CategorySectionQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

func (q CategorySectionQuery) EffectiveLimit() int {
	if q.Limit <= 0 {
		return constants.DefaultLimitCategories
	}
	if q.Limit < constants.DefaultLimitCategories {
		return constants.DefaultLimitCategories
	}
	return q.Limit
}

type CategoryProductsPath struct {
	CategoryID string `param:"category_id" validate:"required"`
}

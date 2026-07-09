package forms

type AdminCategoriesListQuery struct {
	Search string `form:"search"`
	Page   int    `form:"page"`
}

type AdminCategoriesSubcategoriesQuery struct {
	Category string `form:"category" validate:"required"`
}

type AdminCategoriesCreateForm struct {
	Mode          string   `form:"mode" validate:"required"`
	CategoryName  string   `form:"category_name"`
	Category      string   `form:"category"`
	Subcategories []string `form:"subcategories[]"`
}

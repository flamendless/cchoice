package forms

type CategoryPagePath struct {
	Category string `param:"category" validate:"required"`
}

type CategorySubcategoryPagePath struct {
	Category    string `param:"category" validate:"required"`
	Subcategory string `param:"subcategory" validate:"required"`
}

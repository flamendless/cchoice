package forms

type AdminBrandsListQuery struct {
	Search string `form:"search"`
	Status string `form:"status"`
}

type AdminBrandPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminBrandCreateForm struct {
	Name string `form:"name" validate:"required"`
}

type AdminBrandUpdateForm struct {
	Name string `form:"name" validate:"required"`
}

type AdminBrandStatusForm struct {
	Status string `form:"status" validate:"required,brand_status"`
}

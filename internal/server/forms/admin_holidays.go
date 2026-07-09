package forms

type AdminHolidayPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminHolidayCreateForm struct {
	Date string `form:"date" validate:"required"`
	Name string `form:"name" validate:"required"`
	Type string `form:"type" validate:"required"`
}

type AdminHolidayUpdateForm struct {
	Name string `form:"name" validate:"required"`
	Type string `form:"type" validate:"required"`
}

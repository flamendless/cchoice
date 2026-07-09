package forms

type AdminMemoPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminMemoForm struct {
	Title     string   `form:"title" validate:"required"`
	Message   string   `form:"message" validate:"required"`
	FileURL   string   `form:"file_url"`
	Status    string   `form:"status" validate:"required"`
	StartDate string   `form:"start_date" validate:"required"`
	EndDate   string   `form:"end_date" validate:"required"`
	StaffIDs  []string `form:"staff_ids"`
}

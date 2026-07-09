package forms

type AdminQuotationsListQuery struct {
	Search string `form:"search"`
	Page   int    `form:"page"`
}

type AdminQuotationPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminQuotationApproveForm struct {
	AssignedStaffID string `form:"assigned_staff_id" validate:"required"`
	Notes           string `form:"notes"`
}

package forms

type AdminStaffMemoPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminStaffMemoRejectForm struct {
	RejectReason string `form:"reject_reason"`
}

package forms

type AdminSuperuserAttendanceQuery struct {
	StartDate string `form:"date-selector"`
	EndDate   string `form:"date-selector-end"`
	StaffID   string `form:"staff-id"`
}

type AdminSuperuserAttendancePageQuery struct {
	Date string `form:"date"`
}

type AdminSuperuserAttendanceReportForm struct {
	StartDate string `form:"date-selector" validate:"required"`
	EndDate   string `form:"date-selector-end" validate:"required"`
	StaffID   string `form:"staff-id"`
}

type AdminSuperuserAttendanceReportQuery struct {
	Format string `form:"format"`
}

type AdminSuperuserTimeOffPath struct {
	ID string `param:"id" validate:"required"`
}

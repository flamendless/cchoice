package forms

type AdminLogsFilterQuery struct {
	StaffID string `form:"staff-id"`
	Action  string `form:"action"`
	Module  string `form:"module"`
	Page    int    `form:"page"`
}

package forms

type AdminCustomersFilterQuery struct {
	Email  string `form:"email"`
	Type   string `form:"type"`
	Status string `form:"status"`
}

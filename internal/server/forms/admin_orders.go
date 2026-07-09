package forms

type AdminOrdersListQuery struct {
	SearchOrderRef string `form:"search_order_ref"`
	Page           int    `form:"page"`
}

type AdminOrderPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminOrderStatusForm struct {
	Status string `form:"status"`
	Notes  string `form:"notes"`
}

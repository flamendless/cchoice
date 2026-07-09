package forms

type AdminProductInventoriesListQuery struct {
	SearchSerial  string `form:"search_serial"`
	SearchBrand   string `form:"search_brand"`
	ProductStatus string `form:"product_status"`
	StocksIn      string `form:"stocks_in"`
	Page          int    `form:"page"`
}

type AdminProductInventoryPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminProductInventoryUpdateForm struct {
	Qty      string `form:"qty" validate:"required"`
	StocksIn string `form:"stocks_in" validate:"required"`
}

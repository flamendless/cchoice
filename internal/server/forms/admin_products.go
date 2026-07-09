package forms

type AdminProductsCategoryQuery struct {
	Category string `form:"category" validate:"required"`
}

type AdminProductsSerialQuery struct {
	Serial string `form:"serial"`
}

type AdminProductPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminProductsListQuery struct {
	SearchSerial string `form:"search_serial"`
	SearchBrand  string `form:"search_brand"`
	Status       string `form:"status"`
	Page         int    `form:"page"`
}

type AdminProductStatusForm struct {
	Status string `form:"status" validate:"required"`
}

type AdminProductCreateOrUpdateForm struct {
	BrandID           string `form:"brand_id"`
	Category          string `form:"category"`
	Subcategory       string `form:"subcategory"`
	Name              string `form:"name"`
	Serial            string `form:"serial"`
	Description       string `form:"description"`
	Price             string `form:"price"`
	Status            string `form:"status"`
	SpecColours       string `form:"spec_colours"`
	SpecSizes         string `form:"spec_sizes"`
	SpecSegmentation  string `form:"spec_segmentation"`
	SpecPartNumber    string `form:"spec_part_number"`
	SpecPower         string `form:"spec_power"`
	SpecCapacity      string `form:"spec_capacity"`
	SpecScopeOfSupply string `form:"spec_scope_of_supply"`
	SpecWeight        string `form:"spec_weight"`
	SpecWeightUnit    string `form:"spec_weight_unit"`
	StocksIn          string `form:"stocks_in"`
	StocksQty         string `form:"stocks_qty"`
	SalePrice         string `form:"sale_price"`
	SaleStartDate     string `form:"sale_start_date"`
	SaleEndDate       string `form:"sale_end_date"`
}

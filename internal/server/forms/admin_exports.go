package forms

type AdminExportsProductsCountQuery struct {
	Brand  string `form:"brand"`
	Status string `form:"status"`
}

type AdminExportsProductsForm struct {
	Brand         string `form:"brand"`
	Status        string `form:"status"`
	SortColumn    string `form:"sort_column"`
	SortDirection string `form:"sort_direction"`
	Format        string `form:"format"`
}

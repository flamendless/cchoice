package services

type ProductSpecsInput struct {
	Colours, Sizes, Segmentation, PartNumber string
	Power, Capacity, ScopeOfSupply           string
	Weight                                   string
	WeightUnit                               string
}

type CreateProductInput struct {
	Serial, Name, Description string
	BrandID                   string
	Category, Subcategory     string
	Specs                     ProductSpecsInput
	ImagePath                 string
	UnitPriceWithoutVat       int64
	UnitPriceWithVat          int64
}

type UpdateProductInput struct {
	ProductID             string
	BrandID               string
	Category, Subcategory string
	Name, Description     string
	Specs                 ProductSpecsInput
	Status                string
	ImagePath             string
	UnitPriceWithoutVat   int64
	UnitPriceWithVat      int64
}

type ProductForEdit struct {
	ID                          int64
	Serial                      string
	Name                        string
	Description                 string
	BrandID                     int64
	BrandName                   string
	Status                      string
	Category, Subcategory       string
	ProductSpecsID              int64
	UnitPriceWithoutVat         int64
	UnitPriceWithoutVatCurrency string
	UnitPriceWithVat            int64
	UnitPriceWithVatCurrency    string
	Specs                       ProductSpecsInput
}

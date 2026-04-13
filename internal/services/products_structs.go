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

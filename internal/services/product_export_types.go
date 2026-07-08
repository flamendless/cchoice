package services

import "cchoice/internal/enums"

type ProductExportRow struct {
	Brand            string
	Serial           string
	Slug             string
	Status           string
	Category         string
	Subcategory      string
	Name             string
	UnitPriceWithVat string
	SalePriceWithVat string
	SaleStartDate    string
	SaleEndDate      string
	Description      string
	Colours          string
	Sizes            string
	Segmentation     string
	PartNumber       string
	Power            string
	Capacity         string
	Weight           string
	WeightUnit       string
	ScopeOfSupply    string
	StocksIn         string
	StocksQty        string
	ImageURL         string
	ThumbnailURL     string
	CreatedAt        string
	UpdatedAt        string
	LazadaURL        string
	TiktokURL        string
	ShopeeURL        string
}

type ProductImportBlankBehavior int

const (
	ImportBlankSkip ProductImportBlankBehavior = iota
	ImportBlankApply
	ImportReadOnly
)

type ProductImportColumnDef struct {
	Column        string
	BlankBehavior ProductImportBlankBehavior
}

type productExternalPlatformColumn struct {
	Platform enums.ExternalPlatform
	Column   string
}

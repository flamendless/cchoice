package services

import (
	"cchoice/internal/enums"
)

const (
	colRowNumber          = "row number"
	colBrand              = "brand"
	colSerialNumber       = "serial number"
	colProductName        = "product name"
	colSlug               = "slug"
	colStatus             = "status"
	colCategory           = "category"
	colSubcategory        = "subcategory"
	colUnitPriceWithVat   = "unit price with vat"
	colSalePriceWithVat   = "sale price with vat"
	colSaleStartDate      = "sale start date"
	colSaleEndDate        = "sale end date"
	colDescription        = "description"
	colColours            = "colours"
	colSizes              = "sizes"
	colSegmentation       = "segmentation"
	colPartNumber         = "part number"
	colPower              = "power"
	colCapacity           = "capacity"
	colWeight             = "weight"
	colWeightUnit         = "weight unit"
	colScopeOfSupply      = "scope of supply"
	colStocksIn           = "stocks in"
	colStocksQty          = "stocks qty"
	colImageURL           = "product image filename or cdn url"
	colThumbnailURL       = "product image thumbnail filename or cdn url"
	colCreatedAt          = "created at"
	colUpdatedAt          = "updated at"
	colExternalLinkLazada = "external link lazada"
	colExternalLinkTiktok = "external link tiktok"
	colExternalLinkShopee = "external link shopee"
)

var productExportHeaders = []string{
	colRowNumber,
	colBrand,
	colSerialNumber,
	colProductName,
	colSlug,
	colStatus,
	colCategory,
	colSubcategory,
	colUnitPriceWithVat,
	colSalePriceWithVat,
	colSaleStartDate,
	colSaleEndDate,
	colDescription,
	colColours,
	colSizes,
	colSegmentation,
	colPartNumber,
	colPower,
	colCapacity,
	colWeight,
	colWeightUnit,
	colScopeOfSupply,
	colStocksIn,
	colStocksQty,
	colImageURL,
	colThumbnailURL,
	colCreatedAt,
	colUpdatedAt,
	colExternalLinkLazada,
	colExternalLinkTiktok,
	colExternalLinkShopee,
}

var productExportRequiredImportColumns = []string{
	colBrand,
	colSerialNumber,
	colProductName,
}

var productExportRequiredCreateColumns = []string{
	colBrand,
	colSerialNumber,
	colProductName,
	colDescription,
	colCategory,
	colSubcategory,
	colUnitPriceWithVat,
	colColours,
	colSizes,
	colSegmentation,
	colPartNumber,
	colPower,
	colCapacity,
	colWeight,
	colWeightUnit,
	colScopeOfSupply,
	colStocksIn,
	colStocksQty,
}

var productImportColumnDefs = map[string]ProductImportBlankBehavior{
	colRowNumber:          ImportReadOnly,
	colSlug:               ImportReadOnly,
	colCreatedAt:          ImportReadOnly,
	colUpdatedAt:          ImportReadOnly,
	colImageURL:           ImportReadOnly,
	colThumbnailURL:       ImportReadOnly,
	colBrand:              ImportBlankSkip,
	colSerialNumber:       ImportBlankSkip,
	colProductName:        ImportBlankSkip,
	colStatus:             ImportBlankSkip,
	colCategory:           ImportBlankSkip,
	colSubcategory:        ImportBlankSkip,
	colUnitPriceWithVat:   ImportBlankSkip,
	colSalePriceWithVat:   ImportBlankSkip,
	colSaleStartDate:      ImportBlankSkip,
	colSaleEndDate:        ImportBlankSkip,
	colDescription:        ImportBlankSkip,
	colColours:            ImportBlankSkip,
	colSizes:              ImportBlankSkip,
	colSegmentation:       ImportBlankSkip,
	colPartNumber:         ImportBlankSkip,
	colPower:              ImportBlankSkip,
	colCapacity:           ImportBlankSkip,
	colWeight:             ImportBlankSkip,
	colWeightUnit:         ImportBlankSkip,
	colScopeOfSupply:      ImportBlankSkip,
	colStocksIn:           ImportBlankSkip,
	colStocksQty:          ImportBlankSkip,
	colExternalLinkLazada: ImportBlankSkip,
	colExternalLinkTiktok: ImportBlankSkip,
	colExternalLinkShopee: ImportBlankSkip,
}

var productExternalPlatformColumns = []productExternalPlatformColumn{
	{Platform: enums.EXTERNAL_PLATFORM_LAZADA, Column: colExternalLinkLazada},
	{Platform: enums.EXTERNAL_PLATFORM_TIKTOK, Column: colExternalLinkTiktok},
	{Platform: enums.EXTERNAL_PLATFORM_SHOPEE, Column: colExternalLinkShopee},
}

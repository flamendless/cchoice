package services

import "strconv"

func productExportRowToStrings(row ProductExportRow, rowNum int) []string {
	return []string{
		strconv.Itoa(rowNum),
		row.Brand,
		row.Serial,
		row.Name,
		row.Slug,
		row.Status,
		row.Category,
		row.Subcategory,
		row.UnitPriceWithVat,
		row.SalePriceWithVat,
		row.SaleStartDate,
		row.SaleEndDate,
		row.Description,
		row.Colours,
		row.Sizes,
		row.Segmentation,
		row.PartNumber,
		row.Power,
		row.Capacity,
		row.Weight,
		row.WeightUnit,
		row.ScopeOfSupply,
		row.StocksIn,
		row.StocksQty,
		row.ImageURL,
		row.ThumbnailURL,
		row.CreatedAt,
		row.UpdatedAt,
		row.LazadaURL,
		row.TiktokURL,
		row.ShopeeURL,
	}
}

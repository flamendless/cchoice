package models

import (
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
)

type SearchPageData struct {
	Query string
}

type SearchProductsPageData struct {
	Query   string
	Page    int
	HasMore bool
	Products []CategorySectionProduct
}

type SearchRelatedProductsPageData struct {
	Query    string
	Page     int
	HasMore  bool
	Source   string
	Products []CategorySectionProduct
}

func ToProductGridProductsFromSearchRows(
	encoder encode.IEncode,
	getCDNURL CDNURLFunc,
	rows []queries.GetProductsBySearchQueryPaginatedRow,
) []CategorySectionProduct {
	converted := make([]queries.GetProductsByCategoryIDRow, len(rows))
	for i, row := range rows {
		converted[i] = searchPaginatedRowToCategoryRow(row)
	}
	return ToCategorySectionProducts(encoder, getCDNURL, converted)
}

func ToProductGridProductsFromRelatedRows(
	encoder encode.IEncode,
	getCDNURL CDNURLFunc,
	rows []queries.GetRelatedProductsForSearchRow,
) []CategorySectionProduct {
	converted := make([]queries.GetProductsByCategoryIDRow, len(rows))
	for i, row := range rows {
		converted[i] = relatedSearchRowToCategoryRow(row)
	}
	return ToCategorySectionProducts(encoder, getCDNURL, converted)
}

func ToProductGridProductsFromOtherRows(
	encoder encode.IEncode,
	getCDNURL CDNURLFunc,
	rows []queries.GetOtherProductsForSearchRow,
) []CategorySectionProduct {
	converted := make([]queries.GetProductsByCategoryIDRow, len(rows))
	for i, row := range rows {
		converted[i] = otherSearchRowToCategoryRow(row)
	}
	return ToCategorySectionProducts(encoder, getCDNURL, converted)
}

func searchPaginatedRowToCategoryRow(row queries.GetProductsBySearchQueryPaginatedRow) queries.GetProductsByCategoryIDRow {
	return queries.GetProductsByCategoryIDRow{
		ID:                        row.ID,
		Serial:                    row.Serial,
		Slug:                      row.Slug,
		Name:                      row.Name,
		Description:               row.Description,
		UnitPriceWithVat:          row.UnitPriceWithVat,
		UnitPriceWithVatCurrency:  row.UnitPriceWithVatCurrency,
		SalePriceWithVat:          row.SalePriceWithVat,
		SalePriceWithVatCurrency:  row.SalePriceWithVatCurrency,
		IsOnSale:                  row.IsOnSale,
		DiscountType:              row.DiscountType,
		DiscountValue:             row.DiscountValue,
		BrandName:                 row.BrandName,
		ThumbnailPath:             row.ThumbnailPath,
		CdnUrl:                    row.CdnUrl,
		CdnUrlThumbnail:           row.CdnUrlThumbnail,
	}
}

func relatedSearchRowToCategoryRow(row queries.GetRelatedProductsForSearchRow) queries.GetProductsByCategoryIDRow {
	return queries.GetProductsByCategoryIDRow{
		ID:                        row.ID,
		Serial:                    row.Serial,
		Slug:                      row.Slug,
		Name:                      row.Name,
		Description:               row.Description,
		UnitPriceWithVat:          row.UnitPriceWithVat,
		UnitPriceWithVatCurrency:  row.UnitPriceWithVatCurrency,
		SalePriceWithVat:          row.SalePriceWithVat,
		SalePriceWithVatCurrency:  row.SalePriceWithVatCurrency,
		IsOnSale:                  row.IsOnSale,
		DiscountType:              row.DiscountType,
		DiscountValue:             row.DiscountValue,
		BrandName:                 row.BrandName,
		ThumbnailPath:             row.ThumbnailPath,
		CdnUrl:                    row.CdnUrl,
		CdnUrlThumbnail:           row.CdnUrlThumbnail,
	}
}

func otherSearchRowToCategoryRow(row queries.GetOtherProductsForSearchRow) queries.GetProductsByCategoryIDRow {
	return queries.GetProductsByCategoryIDRow{
		ID:                        row.ID,
		Serial:                    row.Serial,
		Slug:                      row.Slug,
		Name:                      row.Name,
		Description:               row.Description,
		UnitPriceWithVat:          row.UnitPriceWithVat,
		UnitPriceWithVatCurrency:  row.UnitPriceWithVatCurrency,
		SalePriceWithVat:          row.SalePriceWithVat,
		SalePriceWithVatCurrency:  row.SalePriceWithVatCurrency,
		IsOnSale:                  row.IsOnSale,
		DiscountType:              row.DiscountType,
		DiscountValue:             row.DiscountValue,
		BrandName:                 row.BrandName,
		ThumbnailPath:             row.ThumbnailPath,
		CdnUrl:                    row.CdnUrl,
		CdnUrlThumbnail:           row.CdnUrlThumbnail,
	}
}

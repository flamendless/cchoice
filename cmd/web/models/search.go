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
		converted[i] = queries.GetProductsByCategoryIDRow(row)
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
		converted[i] = queries.GetProductsByCategoryIDRow(row)
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
		converted[i] = queries.GetProductsByCategoryIDRow(row)
	}
	return ToCategorySectionProducts(encoder, getCDNURL, converted)
}

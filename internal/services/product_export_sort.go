package services

import (
	"slices"
	"strings"

	"cchoice/internal/enums"
)

func sortProductExportRows(
	rows []ProductExportRow,
	sortColumn enums.ProductExportSortColumn,
	sortDirection enums.ProductExportSortDirection,
) {
	if sortColumn == enums.PRODUCT_EXPORT_SORT_COLUMN_UNDEFINED {
		sortColumn = enums.PRODUCT_EXPORT_SORT_COLUMN_UPDATED_AT
	}
	if sortDirection == enums.PRODUCT_EXPORT_SORT_DIRECTION_UNDEFINED {
		sortDirection = enums.PRODUCT_EXPORT_SORT_DIRECTION_DESC
	}

	asc := sortDirection == enums.PRODUCT_EXPORT_SORT_DIRECTION_ASC
	slices.SortFunc(rows, func(a, b ProductExportRow) int {
		var cmp int
		switch sortColumn {
		case enums.PRODUCT_EXPORT_SORT_COLUMN_CREATED_AT:
			cmp = strings.Compare(a.CreatedAt, b.CreatedAt)
		case enums.PRODUCT_EXPORT_SORT_COLUMN_STATUS:
			cmp = strings.Compare(a.Status, b.Status)
		case enums.PRODUCT_EXPORT_SORT_COLUMN_PRODUCT_TITLE:
			cmp = strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		case enums.PRODUCT_EXPORT_SORT_COLUMN_UPDATED_AT:
			fallthrough
		default:
			cmp = strings.Compare(a.UpdatedAt, b.UpdatedAt)
		}
		if cmp == 0 {
			cmp = strings.Compare(a.Serial, b.Serial)
		}
		if asc {
			return cmp
		}
		return -cmp
	})
}

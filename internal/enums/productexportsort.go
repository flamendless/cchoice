package enums

import "strings"

//go:generate go tool stringer -type=ProductExportSortColumn -trimprefix=PRODUCT_EXPORT_SORT_COLUMN_
//go:generate go tool stringer -type=ProductExportSortDirection -trimprefix=PRODUCT_EXPORT_SORT_DIRECTION_

type ProductExportSortColumn int

const (
	PRODUCT_EXPORT_SORT_COLUMN_UNDEFINED ProductExportSortColumn = iota
	PRODUCT_EXPORT_SORT_COLUMN_CREATED_AT
	PRODUCT_EXPORT_SORT_COLUMN_UPDATED_AT
	PRODUCT_EXPORT_SORT_COLUMN_STATUS
	PRODUCT_EXPORT_SORT_COLUMN_PRODUCT_TITLE
)

var AllProductExportSortColumns = []ProductExportSortColumn{
	PRODUCT_EXPORT_SORT_COLUMN_CREATED_AT,
	PRODUCT_EXPORT_SORT_COLUMN_UPDATED_AT,
	PRODUCT_EXPORT_SORT_COLUMN_STATUS,
	PRODUCT_EXPORT_SORT_COLUMN_PRODUCT_TITLE,
}

func (c ProductExportSortColumn) FormValue() string {
	switch c {
	case PRODUCT_EXPORT_SORT_COLUMN_CREATED_AT:
		return "created_at"
	case PRODUCT_EXPORT_SORT_COLUMN_UPDATED_AT:
		return "updated_at"
	case PRODUCT_EXPORT_SORT_COLUMN_STATUS:
		return "status"
	case PRODUCT_EXPORT_SORT_COLUMN_PRODUCT_TITLE:
		return "product_title"
	default:
		return ""
	}
}

func (c ProductExportSortColumn) Label() string {
	switch c {
	case PRODUCT_EXPORT_SORT_COLUMN_CREATED_AT:
		return "Created At"
	case PRODUCT_EXPORT_SORT_COLUMN_UPDATED_AT:
		return "Updated At"
	case PRODUCT_EXPORT_SORT_COLUMN_STATUS:
		return "Status"
	case PRODUCT_EXPORT_SORT_COLUMN_PRODUCT_TITLE:
		return "Product Title"
	default:
		return ""
	}
}

func ParseProductExportSortColumnToEnum(value string) ProductExportSortColumn {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "created_at":
		return PRODUCT_EXPORT_SORT_COLUMN_CREATED_AT
	case "updated_at":
		return PRODUCT_EXPORT_SORT_COLUMN_UPDATED_AT
	case "status":
		return PRODUCT_EXPORT_SORT_COLUMN_STATUS
	case "product_title":
		return PRODUCT_EXPORT_SORT_COLUMN_PRODUCT_TITLE
	default:
		return PRODUCT_EXPORT_SORT_COLUMN_UNDEFINED
	}
}

type ProductExportSortDirection int

const (
	PRODUCT_EXPORT_SORT_DIRECTION_UNDEFINED ProductExportSortDirection = iota
	PRODUCT_EXPORT_SORT_DIRECTION_ASC
	PRODUCT_EXPORT_SORT_DIRECTION_DESC
)

var AllProductExportSortDirections = []ProductExportSortDirection{
	PRODUCT_EXPORT_SORT_DIRECTION_ASC,
	PRODUCT_EXPORT_SORT_DIRECTION_DESC,
}

func (d ProductExportSortDirection) FormValue() string {
	switch d {
	case PRODUCT_EXPORT_SORT_DIRECTION_ASC:
		return "asc"
	case PRODUCT_EXPORT_SORT_DIRECTION_DESC:
		return "desc"
	default:
		return ""
	}
}

func (d ProductExportSortDirection) Label() string {
	switch d {
	case PRODUCT_EXPORT_SORT_DIRECTION_ASC:
		return "Ascending"
	case PRODUCT_EXPORT_SORT_DIRECTION_DESC:
		return "Descending"
	default:
		return ""
	}
}

func ParseProductExportSortDirectionToEnum(value string) ProductExportSortDirection {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "asc":
		return PRODUCT_EXPORT_SORT_DIRECTION_ASC
	case "desc":
		return PRODUCT_EXPORT_SORT_DIRECTION_DESC
	default:
		return PRODUCT_EXPORT_SORT_DIRECTION_UNDEFINED
	}
}

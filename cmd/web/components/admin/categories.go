package components

import (
	"net/url"
	"strconv"

	"cchoice/internal/utils"
)

func categorySubcategoriesURL(category string) string {
	return utils.URL("/admin/categories/subcategories?category=" + url.QueryEscape(category))
}

func categoryRowID(index int) string {
	return "category-sub-" + strconvItoa(index)
}

func categorySubCellID(index int) string {
	return "category-sub-cell-" + strconvItoa(index)
}

func strconvItoa(n int) string {
	return strconv.Itoa(n)
}

package components

import (
	"slices"

	"cchoice/internal/enums"
)

type ExportSubcard struct {
	Title       string
	Description string
	ModalURL    string
}

var exportSubcardsWithRoles = []struct {
	Card        ExportSubcard
	AllowedRole enums.StaffRole
}{
	{
		Card: ExportSubcard{
			Title:       "Products",
			Description: "Export products to CSV",
			ModalURL:    "/admin/exports/products/modal",
		},
		AllowedRole: enums.STAFF_ROLE_EXPORTS_PRODUCTS,
	},
}

func getExportSubcards(isSuperuser bool, roles []enums.StaffRole) []ExportSubcard {
	if isSuperuser {
		return []ExportSubcard{
			{
				Title:       "Products",
				Description: "Export products to CSV",
				ModalURL:    "/admin/exports/products/modal",
			},
		}
	}

	result := make([]ExportSubcard, 0, len(exportSubcardsWithRoles))
	for _, item := range exportSubcardsWithRoles {
		if slices.Contains(roles, item.AllowedRole) {
			result = append(result, item.Card)
		}
	}
	return result
}

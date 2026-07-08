package components

import (
	"slices"

	"cchoice/internal/enums"
)

type ImportSubcard struct {
	Title       string
	Description string
	ModalURL    string
}

var importSubcardsWithRoles = []struct {
	Card        ImportSubcard
	AllowedRole enums.StaffRole
}{
	{
		Card: ImportSubcard{
			Title:       "Products",
			Description: "Bulk upload products from CSV or XLSX",
			ModalURL:    "/admin/imports/products/modal",
		},
		AllowedRole: enums.STAFF_ROLE_EDIT_PRODUCTS,
	},
}

func getImportSubcards(isSuperuser bool, roles []enums.StaffRole) []ImportSubcard {
	if isSuperuser {
		return []ImportSubcard{
			{
				Title:       "Products",
				Description: "Bulk upload products from CSV or XLSX",
				ModalURL:    "/admin/imports/products/modal",
			},
		}
	}

	result := make([]ImportSubcard, 0, len(importSubcardsWithRoles))
	for _, item := range importSubcardsWithRoles {
		if slices.Contains(roles, item.AllowedRole) {
			result = append(result, item.Card)
		}
	}
	return result
}

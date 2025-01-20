package models

import "cchoice/internal/database/queries"

type HeaderRowText struct {
	Label string
	URL   string
}

type FooterRowText struct {
	Label string
	URL   string
}

type CategorySidePanelText struct {
	Label string
	URL   string
}

type CategorySection struct {
	ID    string
	Label string
}

type CategorySectionProduct queries.GetProductsByCategoryIDRow

type CategorySectionProducts struct {
	ID          string
	Category    string
	Subcategory string
	Products    []CategorySectionProduct
}

func ToCategorySectionProducts[T queries.GetProductsByCategoryIDRow](data []T) []CategorySectionProduct {
	res := make([]CategorySectionProduct, 0, len(data))
	for _, d := range data {
		res = append(res, CategorySectionProduct(d))
	}
	return res
}

package models

import (
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/utils"
	"fmt"

	"github.com/Rhymond/go-money"
)

type HeaderRowText struct {
	Label string
	URL   string
}

type FooterRowText struct {
	Label    string
	URL      string
	Hideable bool
}

type CategorySidePanelText struct {
	Label          string
	URL            string
	ScrollTargetID string
}

type BrandSidePanelText struct {
	Label   string
	URL     string
	BrandID string
}

type CategorySection struct {
	ID    string
	Label string
}

type Subcategory struct {
	CategoryID string
	Label      string
}

type GroupedCategorySection struct {
	Label          string
	ScrollTargetID string
	Subcategories  []Subcategory
}

type CategorySectionProduct struct {
	queries.GetProductsByCategoryIDRow
	ProductID          string
	CDNURL             string
	CDNURL1280         string
	PriceDisplay       string
	DiscountPercentage string
}

type CategorySectionProducts struct {
	ID          string
	Category    string
	Subcategory string
	Products    []CategorySectionProduct
}

type CDNURLFunc func(path string) string

func ToCategorySectionProducts[T queries.GetProductsByCategoryIDRow](
	encoder encode.IEncode,
	getCDNURL CDNURLFunc,
	data []T,
) []CategorySectionProduct {
	res := make([]CategorySectionProduct, 0, len(data))
	for _, d := range data {
		r := queries.GetProductsByCategoryIDRow(d)
		var price *money.Money
		var discountPercentage string
		if r.IsOnSale == 1 {
			price = utils.NewMoney(r.SalePriceWithVat.Int64, r.SalePriceWithVatCurrency.String)
			discount := ((r.UnitPriceWithVat - r.SalePriceWithVat.Int64) * 100.0) / r.UnitPriceWithVat
			discountPercentage = fmt.Sprintf("%d%%", discount)
		} else {
			price = utils.NewMoney(r.UnitPriceWithVat, r.UnitPriceWithVatCurrency)
		}
		res = append(res, CategorySectionProduct{
			GetProductsByCategoryIDRow: r,
			ProductID:                  encoder.Encode(r.ID),
			CDNURL:                     getCDNURL(r.ThumbnailPath),
			CDNURL1280:                 getCDNURL(constants.ToPath1280(r.ThumbnailPath)),
			PriceDisplay:               price.Display(),
			DiscountPercentage:         discountPercentage,
		})
	}
	return res
}

type SearchResultProduct struct {
	queries.GetProductsBySearchQueryRow
	ProductID    string
	CDNURL       string
	CDNURL1280   string
	PriceDisplay string
}

func ToSearchResultProduct[T queries.GetProductsBySearchQueryRow](
	encoder encode.IEncode,
	getCDNURL CDNURLFunc,
	data T,
) SearchResultProduct {
	r := queries.GetProductsBySearchQueryRow(data)
	price := utils.NewMoney(r.UnitPriceWithVat, r.UnitPriceWithVatCurrency)
	return SearchResultProduct{
		GetProductsBySearchQueryRow: r,
		ProductID:                   encoder.Encode(r.ID),
		CDNURL:                      getCDNURL(r.ThumbnailPath),
		CDNURL1280:                  getCDNURL(constants.ToPath1280(r.ThumbnailPath)),
		PriceDisplay:                price.Display(),
	}
}

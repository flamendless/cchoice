package models

import (
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/utils"
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
	OrigPriceDisplay   string
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
		origPrice, discountedPrice, discountPercentage := utils.GetOrigAndDiscounted(
			r.IsOnSale,
			r.UnitPriceWithVat,
			r.UnitPriceWithVatCurrency,
			r.SalePriceWithVat,
			r.SalePriceWithVatCurrency,
		)

		res = append(res, CategorySectionProduct{
			GetProductsByCategoryIDRow: r,
			ProductID:                  encoder.Encode(r.ID),
			CDNURL:                     getCDNURL(r.ThumbnailPath),
			CDNURL1280:                 getCDNURL(constants.ToPath1280(r.ThumbnailPath)),
			OrigPriceDisplay:           origPrice.Display(),
			PriceDisplay:               discountedPrice.Display(),
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

type AdminStaffProfile struct {
	FullName         string
	Birthdate        string
	DateHired        string
	Position         string
	Email            string
	MobileNo         string
	ScheduledTimeIn  string
	ScheduledTimeOut string
	SelectedDate     string
	CurrentDate      string
	CurrentTime      string
	HasTimeIn        bool
	HasTimeOut       bool
	CanTimeIn        bool
	CanTimeOut       bool
	RequireInShop    bool
	MyAttendance     *AdminStaffAttendance
	InShop           *bool
	LocationDisplay  string
	UserType         enums.StaffUserType
}

type AdminStaffAttendance struct {
	StaffID          int64
	FullName         string
	TimeIn           string
	TimeOut          string
	ScheduledTimeIn  string
	ScheduledTimeOut string
	TimeInStatus     enums.TimeInStatus
	TimeOutStatus    enums.TimeOutStatus
	Duration         string
	DurationColor    string
	InShop           bool
}

type AdminSuperuserPage struct {
	FullName     string
	CurrentDate  string
	CurrentTime  string
	SelectedDate string
	Attendances  []AdminStaffAttendance
}

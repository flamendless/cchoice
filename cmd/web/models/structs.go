package models

import (
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/utils"
	"time"
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
	FirstName        string
	MiddleName       string
	LastName         string
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
	CanLunchBreakIn  bool
	CanLunchBreakOut bool
	RequireInShop    bool
	MyAttendance     *Attendance
	InShop           *bool
	OutShop          *bool
	LocationDisplay  string
	DistanceMeters   float64
	Lat              float64
	Lng              float64
	UserType         enums.StaffUserType
}

type AttendanceStat struct {
	In            string
	Out           string
	InStatus      enums.TimeInStatus
	OutStatus     enums.TimeOutStatus
	Duration      string
	DurationColor string
	InLate        time.Duration
	InShop        bool
	OutShop       bool
	InLocation    string
	OutLocation   string
	InDeviceInfo  string
	OutDeviceInfo string
}

type Attendance struct {
	StaffID          string
	FullName         string
	Date             string
	ScheduledTimeIn  string
	ScheduledTimeOut string
	Attendance       AttendanceStat
	LunchBreak       AttendanceStat
}

type AdminSuperuserPage struct {
	FullName     string
	CurrentDate  string
	CurrentTime  string
	SelectedDate string
	Attendances  []Attendance
}

type AdminBrand struct {
	ID   string
	Name string
}

type AdminCategory struct {
	Category      string
	Subcategories []string
}

type AdminSubcategory struct {
	Category    string
	Subcategory string
}

type AdminProductForm struct {
	Brands         []AdminBrand
	Categories     []AdminCategory
	CategoriesJSON string
	Subcategories  []AdminSubcategory
	VATPercentage  string
}

type AdminProductSpecsForm struct {
	Colours       string
	Sizes         string
	Segmentation  string
	PartNumber    string
	Power         string
	Capacity      string
	ScopeOfSupply string
}

type AdminProductListItem struct {
	ID          string
	Name        string
	Serial      string
	Description string
	Brand       string
	Status      enums.ProductStatus
	ImagePath   string
	CreatedAt   string
	UpdatedAt   string
}

type StaffTimeOff struct {
	ID          string
	StaffID     string
	FullName    string
	Type        enums.TimeOff
	StartDate   string
	EndDate     string
	Description string
	CreatedAt   string
	Approved    bool
	ApprovedBy  string
	ApprovedAt  string
}

type Staff struct {
	ID       string
	FullName string
}

type StaffLog struct {
	ID         string
	StaffID    string
	FullName   string
	FirstName  string
	MiddleName string
	LastName   string
	CreatedAt  string
	Action     string
	Module     string
	Result     string
}

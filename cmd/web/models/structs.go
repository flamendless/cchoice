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
	ProductID          string
	CDNURL             string
	CDNURL1280         string
	OrigPriceDisplay   string
	PriceDisplay       string
	DiscountPercentage string
	queries.GetProductsByCategoryIDRow
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
			CDNURL:                     r.CdnUrlThumbnail.String,
			CDNURL1280:                 r.CdnUrl.String,
			OrigPriceDisplay:           origPrice.Display(),
			PriceDisplay:               discountedPrice.Display(),
			DiscountPercentage:         discountPercentage,
		})
	}
	return res
}

type SearchResultProduct struct {
	ProductID    string
	CDNURL       string
	CDNURL1280   string
	PriceDisplay string
	queries.GetProductsBySearchQueryRow
}

func ToSearchResultProduct[T queries.GetProductsBySearchQueryRow](
	encoder encode.IEncode,
	getCDNURL CDNURLFunc,
	data T,
) SearchResultProduct {
	r := queries.GetProductsBySearchQueryRow(data)
	price := utils.NewMoney(r.UnitPriceWithVat, r.UnitPriceWithVatCurrency)

	cdnURL := r.CdnUrl.String
	if cdnURL == "" {
		cdnURL = getCDNURL(r.ThumbnailPath)
	}
	cdnURLThumbnail := r.CdnUrlThumbnail.String
	if cdnURLThumbnail == "" {
		cdnURLThumbnail = getCDNURL(constants.ToPath1280(r.ThumbnailPath))
	}

	return SearchResultProduct{
		GetProductsBySearchQueryRow: r,
		ProductID:                   encoder.Encode(r.ID),
		CDNURL:                      cdnURL,
		CDNURL1280:                  cdnURLThumbnail,
		PriceDisplay:                price.Display(),
	}
}

type AdminStaffProfile struct {
	MyAttendance     *Attendance
	InShop           *bool
	OutShop          *bool
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
	LocationDisplay  string
	DistanceMeters   float64
	Lat              float64
	Lng              float64
	UserType         enums.StaffUserType
	HasTimeIn        bool
	HasTimeOut       bool
	CanTimeIn        bool
	CanTimeOut       bool
	CanLunchBreakIn  bool
	CanLunchBreakOut bool
	RequireInShop    bool
}

type CustomerProfile struct {
	FullName     string
	FirstName    string
	MiddleName   string
	LastName     string
	Birthdate    string
	Sex          string
	Email        string
	MobileNo     string
	CompanyName  string
	CustomerType enums.CustomerType
	Status       enums.CustomerStatus
}

type AttendanceStat struct {
	In            string
	Out           string
	Duration      string
	DurationColor string
	InLocation    string
	OutLocation   string
	InDeviceInfo  string
	OutDeviceInfo string
	InStatus      enums.TimeInStatus
	OutStatus     enums.TimeOutStatus
	InLate        time.Duration
	Undertime     time.Duration
	EarlyIn       time.Duration
	InShop        bool
	OutShop       bool
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
	CategoriesJSON string
	VATPercentage  string
	Brands         []AdminBrand
	Categories     []AdminCategory
	Subcategories  []AdminSubcategory
}

type AdminProductSpecsForm struct {
	Colours       string
	Sizes         string
	Segmentation  string
	PartNumber    string
	Power         string
	Capacity      string
	ScopeOfSupply string
	Weight        string
	WeightUnit    string
}

type AdminProductListItem struct {
	ID            string
	Name          string
	Serial        string
	Description   string
	Brand         string
	ThumbnailPath string
	CDNURL        string
	CDNURL1280    string
	CreatedAt     string
	UpdatedAt     string
	Colours       string
	Sizes         string
	Segmentation  string
	PartNumber    string
	Power         string
	Capacity      string
	ScopeOfSupply string
	Weight        string
	WeightUnit    string
	Status        enums.ProductStatus
}

type StaffTimeOff struct {
	ID          string
	StaffID     string
	FullName    string
	StartDate   string
	EndDate     string
	Description string
	CreatedAt   string
	ApprovedBy  string
	ApprovedAt  string
	Type        enums.TimeOff
	Approved    bool
}

type Staff struct {
	ID       string
	FullName string
}

type AdminStaffListItem struct {
	ID       string
	FullName string
	Position string
	Email    string
	MobileNo string
	Roles    []enums.StaffRole
	UserType enums.StaffUserType
}

type AdminHolidayListItem struct {
	ID   string
	Date string
	Name string
	Type enums.HolidayType
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

type AdminCustomerListItem struct {
	ID           string
	Email        string
	FirstName    string
	MiddleName   string
	LastName     string
	Birthdate    string
	Sex          string
	CompanyName  string
	CreatedAt    string
	CustomerType enums.CustomerType
	IsVerified   enums.CustomerStatus
}

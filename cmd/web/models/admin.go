package models

import (
	"cchoice/internal/enums"

	"github.com/a-h/templ"
)

type StaffCard struct {
	Link        string
	Title       string
	Description string
	Icon        templ.Component
}

type AdminPromoListItem struct {
	ID          string
	Title       string
	Description string
	MediaURL    string
	StartDate   string
	EndDate     string
	Type        enums.PromoType
	Status      enums.PromoStatus
	BannerOnly  bool
	Priority    int64
	CreatedAt   string
}

type AdminMemoListItem struct {
	ID            string
	Title         string
	Message       string
	FileURL       string
	Status        enums.MemoStatus
	StartDate     string
	EndDate       string
	CreatedByID   string
	CreatedByName string
	CreatedAt     string
	EmailsSentAt  string
	RecipientIDs  []string
}

type AdminMemoRecipientRow struct {
	StaffID      string
	StaffName    string
	Email        string
	Position     string
	UserType     enums.StaffUserType
	ActionStatus enums.MemoStaffActionStatus
	RejectReason string
	AcceptedAt   string
	RejectedAt   string
}

type StaffMemoCard struct {
	ID      string
	Title   string
	Message string
	FileURL string
}

type AdminCategoryListItem struct {
	Category           string
	SubcategoriesCount int64
	ProductsCount      int64
}

type AdminSubcategoryRow struct {
	ID          string
	Subcategory string
	Promoted    bool
}

type AdminOrderListItem struct {
	ID             string
	OrderReference string
	Status         enums.OrderStatus
	IsPaid         bool
	CreatedAt      string
	UpdatedAt      string
	EarnedCPoints  string
}

type AdminOrderLineItem struct {
	ThumbnailURL string
	Name         string
	Serial       string
	UnitPrice    string
	Quantity     int64
	TotalPrice   string
}

type AdminOrderCustomerInfo struct {
	Name  string
	Email string
	Phone string
}

type AdminOrderInfo struct {
	OrderReference string
	Status         enums.OrderStatus
	Notes          string
	Remarks        string
	CreatedAt      string
	UpdatedAt      string
	EarnedCPoints  string
}

type AdminOrderPaymentInfo struct {
	Gateway         string
	Status          string
	ReferenceNumber string
	PaymentMethod   string
	TotalAmount     string
	PaidAt          string
	Description     string
	MetadataNotes   string
	MetadataRemarks string
	CustomerNumber  string
}

type AdminOrderAddressInfo struct {
	Line1            string
	Line2            string
	City             string
	State            string
	PostalCode       string
	Country          string
	FormattedAddress string
}

type AdminOrderShippingInfo struct {
	AdminOrderAddressInfo
	Service        string
	OrderID        string
	TrackingNumber string
	ETA            string
}

type AdminOrderAmountSummary struct {
	Subtotal string
	Shipping string
	Discount string
	Total    string
}

type AdminOrderDetails struct {
	Order    AdminOrderInfo
	Payment  AdminOrderPaymentInfo
	Shipping AdminOrderShippingInfo
	Billing  AdminOrderAddressInfo
	Customer AdminOrderCustomerInfo
	Summary  AdminOrderAmountSummary
	Lines    []AdminOrderLineItem
}

type AdminOrderManageModalData struct {
	ID             string
	OrderReference string
	CurrentStatus  enums.OrderStatus
	CanUpdateStatus bool
}

type AdminOrderStatusHistoryEntry struct {
	FromStatus string
	ToStatus   string
	StaffName  string
	Notes      string
	CreatedAt  string
}

type AdminOrderTrackModalData struct {
	ID             string
	OrderReference string
	CurrentStatus  enums.OrderStatus
	History        []AdminOrderStatusHistoryEntry
	FlowSteps      []enums.OrderStatus
}

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
}

type AdminOrderLineItem struct {
	Name        string
	Serial      string
	Description string
	UnitPrice   string
	Quantity    int64
	TotalPrice  string
}

type AdminOrderCustomerInfo struct {
	Name    string
	Email   string
	Phone   string
	Address string
}

type AdminOrderDetails struct {
	Customer AdminOrderCustomerInfo
	Lines    []AdminOrderLineItem
}

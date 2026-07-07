package services

import "cchoice/internal/enums"

type QuotationAdminListItem struct {
	ID           int64
	CustomerName string
	Status       enums.QuotationStatus
	AssignedTo   string
	TotalItems   int64
	TotalDisplay string
	SubmittedAt  string
	UpdatedAt    string
}

type QuotationCustomerListItem struct {
	ID           int64
	Status       enums.QuotationStatus
	TotalItems   int64
	TotalDisplay string
	SubmittedAt  string
}

type QuotationAdminLineItem struct {
	BrandName     string
	ProductSerial string
	Quantity      int64
	TotalPrice    string
	TotalDiscount string
}

type QuotationStatusHistoryEntry struct {
	FromStatus string
	ToStatus   string
	StaffName  string
	Notes      string
	CreatedAt  string
}

type QuotationAdminTrackData struct {
	ID            int64
	CurrentStatus enums.QuotationStatus
	History       []QuotationStatusHistoryEntry
	FlowSteps     []enums.QuotationStatus
}

type QuotationCustomerDetailData struct {
	ID           int64
	Status       enums.QuotationStatus
	SubmittedAt  string
	UpdatedAt    string
	Lines        []QuotationAdminLineItem
	TotalItems   int64
	TotalPrice   string
	TotalDiscounts string
	Total        string
	Track        QuotationAdminTrackData
}

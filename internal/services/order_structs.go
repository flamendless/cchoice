package services

import "cchoice/internal/enums"

type OrderAdminListItem struct {
	ID             int64
	OrderReference string
	Status         enums.OrderStatus
	IsPaid         bool
	CreatedAt      string
	UpdatedAt      string
}

type OrderAdminCustomerInfo struct {
	Name  string
	Email string
	Phone string
}

type OrderAdminInfo struct {
	OrderReference string
	Status         enums.OrderStatus
	Notes          string
	Remarks        string
	CreatedAt      string
	UpdatedAt      string
}

type OrderAdminPaymentInfo struct {
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

type OrderAdminAddressInfo struct {
	Line1            string
	Line2            string
	City             string
	State            string
	PostalCode       string
	Country          string
	FormattedAddress string
}

type OrderAdminShippingInfo struct {
	OrderAdminAddressInfo
	Service        string
	OrderID        string
	TrackingNumber string
	ETA            string
}

type OrderAdminAmountSummary struct {
	Subtotal string
	Shipping string
	Discount string
	Total    string
}

type OrderAdminLineItem struct {
	ThumbnailURL string
	Name         string
	Serial       string
	UnitPrice    string
	Quantity     int64
	TotalPrice   string
}

type OrderAdminDetails struct {
	Order    OrderAdminInfo
	Payment  OrderAdminPaymentInfo
	Shipping OrderAdminShippingInfo
	Billing  OrderAdminAddressInfo
	Customer OrderAdminCustomerInfo
	Summary  OrderAdminAmountSummary
	Lines    []OrderAdminLineItem
}

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
	Name    string
	Email   string
	Phone   string
	Address string
}

type OrderAdminLineItem struct {
	Name        string
	Serial      string
	Description string
	UnitPrice   string
	Quantity    int64
	TotalPrice  string
}

type OrderAdminDetails struct {
	Customer OrderAdminCustomerInfo
	Lines    []OrderAdminLineItem
}

package scanner

import "time"

type LineItem struct {
	Name     string
	Quantity string
	Price    string
	Subtotal string
}

type ReceiptData struct {
	MerchantName    string
	MerchantAddress string
	MerchantPhone   string
	MerchantTIN     string
	MerchantProp    string
	ReceiptType     string
	ReceiptNumber   string
	Date            string
	Time            string
	SoldTo          string
	CustomerTIN     string
	CustomerAddress string
	Subtotal        string
	Tax             string
	Total           string
	PaymentMethod   string
	RawText         string
	Currency        string
	VATInclusive    string
	LessVAT         string
	LessWithholding string
	AmountNetOfVAT  string
	AddVAT          string
	Items           []LineItem
}

type ScanResult struct {
	ScannedAt time.Time
	Error     error
	Data      *ReceiptData
	ImagePath string
	Success   bool
}

type IReceiptScanner interface {
	ScanReceipt(imagePath string) (*ReceiptData, error)
	Close() error
}

package models

import "github.com/Rhymond/go-money"

type CheckoutLine struct {
	ID            string
	CheckoutID    string
	ProductID     string
	Name          string
	Price         money.Money
	Total         money.Money
	BrandName     string
	Quantity      int64
	ThumbnailPath string
	ThumbnailData string
}

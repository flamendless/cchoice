package models

import "github.com/Rhymond/go-money"

type CheckoutLine struct {
	Price           money.Money
	DiscountedPrice money.Money
	Total           money.Money
	ID              string
	CheckoutID      string
	ProductID       string
	Name            string
	BrandName       string
	ThumbnailPath   string
	ThumbnailData   string
	Quantity        int64
}

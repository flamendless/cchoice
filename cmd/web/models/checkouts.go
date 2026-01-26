package models

import "github.com/Rhymond/go-money"

type CheckoutLine struct {
	OrigPrice          money.Money
	Price              money.Money
	Total              money.Money
	ID                 string
	CheckoutID         string
	ProductID          string
	Name               string
	BrandName          string
	ThumbnailPath      string
	CDNURL             string
	CDNURL1280         string
	Quantity           int64
	WeightKg           float64
	WeightDisplay      string
	Checked            bool
	DiscountPercentage string
}

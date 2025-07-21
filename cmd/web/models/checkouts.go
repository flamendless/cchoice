package models

type CheckoutLine struct {
	ID         string
	CheckoutID string
	ProductID  string
	Name       string
	BrandName  string
	Quantity   int64
}

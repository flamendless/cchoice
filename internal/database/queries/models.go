// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package queries

import (
	"database/sql"
	"time"
)

type TblBrand struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type TblBrandImage struct {
	ID        int64
	BrandID   int64
	Path      string
	IsMain    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type TblCheckout struct {
	ID        int64
	SessionID string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TblCheckoutLine struct {
	ID          int64
	CheckoutID  int64
	ProductID   int64
	Name        string
	Serial      string
	Description string
	Amount      int64
	Currency    string
	Quantity    int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TblCheckoutPayment struct {
	ID                     string
	Gateway                string
	CheckoutID             int64
	Status                 string
	Description            string
	TotalAmount            int64
	CheckoutUrl            string
	ClientKey              string
	ReferenceNumber        string
	PaymentStatus          string
	PaymentMethodType      string
	PaidAt                 time.Time
	MetadataRemarks        string
	MetadataNotes          string
	MetadataCustomerNumber string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type TblProduct struct {
	ID                          int64
	Serial                      string
	Name                        string
	Description                 sql.NullString
	BrandID                     int64
	Status                      string
	ProductSpecsID              sql.NullInt64
	UnitPriceWithoutVat         int64
	UnitPriceWithVat            int64
	UnitPriceWithoutVatCurrency string
	UnitPriceWithVatCurrency    string
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	DeletedAt                   time.Time
}

type TblProductCategory struct {
	ID                 int64
	Category           sql.NullString
	Subcategory        sql.NullString
	PromotedAtHomepage sql.NullBool
}

type TblProductImage struct {
	ID        int64
	ProductID int64
	Path      string
	Thumbnail sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type TblProductSpec struct {
	ID            int64
	Colours       sql.NullString
	Sizes         sql.NullString
	Segmentation  sql.NullString
	PartNumber    sql.NullString
	Power         sql.NullString
	Capacity      sql.NullString
	ScopeOfSupply sql.NullString
}

type TblProductsCategory struct {
	ID         int64
	CategoryID int64
	ProductID  int64
}

type TblProductsFt struct {
	Serial string
	Name   string
}

type TblSetting struct {
	ID    int64
	Name  string
	Value string
}

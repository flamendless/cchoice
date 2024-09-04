// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package cchoice_db

import (
	"database/sql"
	"time"
)

type TblAuth struct {
	ID            int64
	UserID        int64
	Token         string
	OtpEnabled    bool
	OtpSecret     sql.NullString
	RecoveryCodes sql.NullString
}

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
	ID          int64
	ProductID   int64
	Category    sql.NullString
	Subcategory sql.NullString
}

type TblProductImage struct {
	ID        int64
	ProductID int64
	Path      string
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

type TblUser struct {
	ID         int64
	FirstName  string
	MiddleName string
	LastName   string
	Email      string
	Password   string
	MobileNo   string
	UserType   string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  time.Time
}

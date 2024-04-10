// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package cchoice_db

import (
	"database/sql"
)

type Product struct {
	ID                          int64
	Name                        string
	Description                 sql.NullString
	Status                      sql.NullString
	ProductCategoryID           sql.NullInt64
	Colours                     sql.NullString
	Sizes                       sql.NullString
	Segmentation                sql.NullString
	UnitPriceWithoutVat         int64
	UnitPriceWithVat            int64
	UnitPriceWithoutVatCurrency string
	UnitPriceWithVatCurrency    string
	CreatedAt                   string
	UpdatedAt                   string
	DeletedAt                   string
}

type ProductCategory struct {
	ID          int64
	Category    sql.NullString
	Subcategory sql.NullString
}

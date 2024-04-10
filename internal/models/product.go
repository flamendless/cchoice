package models

import (
	cchoice_db "cchoice/db"
	"cchoice/internal"
	"cchoice/internal/constants"
	"cchoice/internal/logs"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/gosimple/slug"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type Product struct {
	ID                  int64
	Serial              string
	Name                string
	Description         string
	Status              ProductStatus
	Category            string
	Subcategory         string
	Colours             string
	Sizes               string
	Segmentation        string
	UnitPriceWithoutVat *money.Money
	UnitPriceWithVat    *money.Money
}

func (product *Product) PostProcess() {
	product.Category = slug.Make(product.Category)
	product.Subcategory = slug.Make(product.Subcategory)
}

func (product *Product) Print() {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("ID: %d\n", product.ID))
	builder.WriteString(fmt.Sprintf("Serial: %s\n", product.Serial))
	builder.WriteString(fmt.Sprintf("Name: %s\n", product.Name))
	builder.WriteString(fmt.Sprintf("Description: %s\n", product.Description))
	builder.WriteString(fmt.Sprintf("Product Status: %s\n", &product.Status))

	builder.WriteString(fmt.Sprintf("Category: %s\n", product.Category))
	builder.WriteString(fmt.Sprintf("Subcategory: %s\n", product.Subcategory))

	builder.WriteString(fmt.Sprintf("Colours: %s\n", product.Colours))
	builder.WriteString(fmt.Sprintf("Sizes: %s\n", product.Sizes))
	builder.WriteString(fmt.Sprintf("Segmentation: %s\n", product.Segmentation))

	builder.WriteString(fmt.Sprintf("Unit Price w/o VAT: %s\n", product.UnitPriceWithoutVat.Display()))
	builder.WriteString(fmt.Sprintf("Unit Price w VAT: %s\n", product.UnitPriceWithVat.Display()))

	fmt.Println(builder.String())
}

func (product *Product) Duplicate() *Product {
	newProduct := Product{
		Name:                product.Name,
		Description:         product.Description,
		Category:            product.Category,
		Subcategory:         product.Subcategory,
		Colours:             product.Colours,
		Sizes:               product.Sizes,
		Segmentation:        product.Segmentation,
		UnitPriceWithoutVat: product.UnitPriceWithoutVat,
		UnitPriceWithVat:    product.UnitPriceWithVat,
	}
	return &newProduct
}

func (product *Product) InsertToDB(appCtx *internal.AppContext) (int64, error) {
	ctx := context.Background()
	queries := appCtx.Queries

	var categoryID int64

	if product.Category != "" {
		existingProductCategory, err := queries.GetProductCategoryByCategoryAndSubcategory(
			ctx,
			cchoice_db.GetProductCategoryByCategoryAndSubcategoryParams{
				Category: sql.NullString{
					String: product.Category,
					Valid:  true,
				},
				Subcategory: sql.NullString{
					String: product.Subcategory,
					Valid:  true,
				},
			},
		)
		if err != nil {
			existingProductCategory, err = queries.CreateProductCategory(
				ctx,
				cchoice_db.CreateProductCategoryParams{
					Category: sql.NullString{
						String: product.Category,
						Valid:  true,
					},
					Subcategory: sql.NullString{
						String: product.Subcategory,
						Valid:  product.Subcategory != "",
					},
				},
			)
			if err != nil {
				logs.Log().Warn(
					"Insert product category",
					zap.Error(err),
				)
			}
		}
		categoryID = existingProductCategory.ID
	}

	now := time.Now().UTC().String()

	insertedProduct, err := queries.CreateProduct(
		ctx,
		cchoice_db.CreateProductParams{
			Name: product.Name,
			Description: sql.NullString{
				String: product.Description,
				Valid:  true,
			},
			Status: sql.NullString{
				String: product.Status.String(),
				Valid:  true,
			},
			ProductCategoryID: sql.NullInt64{
				Int64: categoryID,
				Valid: categoryID != 0,
			},
			Colours: sql.NullString{
				String: product.Colours,
				Valid:  true,
			},
			Sizes: sql.NullString{
				String: product.Sizes,
				Valid:  true,
			},
			Segmentation: sql.NullString{
				String: product.Segmentation,
				Valid:  true,
			},

			UnitPriceWithoutVat: product.UnitPriceWithoutVat.Amount() * 100,
			UnitPriceWithVat:    product.UnitPriceWithVat.Amount() * 100,

			UnitPriceWithoutVatCurrency: product.UnitPriceWithoutVat.Currency().Code,
			UnitPriceWithVatCurrency:    product.UnitPriceWithVat.Currency().Code,

			CreatedAt: now,
			UpdatedAt: now,
			DeletedAt: constants.DT_BEGINNING.String(),
		},
	)
	if err != nil {
		return 0, err
	}

	return insertedProduct.ID, nil
}

func DBRowToProduct(row *cchoice_db.GetProductRow) *Product {
	dbp := &Product{
		ID:           row.ID,
		Name:         row.Name,
		Description:  row.Description.String,
		Status:       ParseProductStatusEnum(row.Status.String),
		Category:     row.Category.String,
		Subcategory:  row.Subcategory.String,
		Colours:      row.Colours.String,
		Sizes:        row.Sizes.String,
		Segmentation: row.Segmentation.String,
	}

	unitPriceWithoutVat := decimal.NewFromInt(row.UnitPriceWithoutVat / 100)
	unitPriceWithVat := decimal.NewFromInt(row.UnitPriceWithVat / 100)

	dbp.UnitPriceWithoutVat = money.New(
		unitPriceWithoutVat.CoefficientInt64(),
		row.UnitPriceWithoutVatCurrency,
	)
	dbp.UnitPriceWithVat = money.New(
		unitPriceWithVat.CoefficientInt64(),
		row.UnitPriceWithVatCurrency,
	)

	return dbp
}

package models

import (
	cchoice_db "cchoice/cchoice_db"
	"cchoice/internal"
	"cchoice/internal/constants"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
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
	Brand               string
	Status              ProductStatus
	Category            string
	Subcategory         string
	Colours             string
	Sizes               string
	Segmentation        string
	UnitPriceWithoutVat *money.Money
	UnitPriceWithVat    *money.Money
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           time.Time
}

func (product *Product) PostProcess(rowIdx int) {
	brandInitials := utils.GetInitials(product.Brand)
	nameInitials := utils.GetInitials(product.Name)
	product.Serial = fmt.Sprintf("%s-%s-%d", brandInitials, nameInitials, rowIdx)
	product.Category = slug.Make(product.Category)
	product.Subcategory = slug.Make(product.Subcategory)
}

func (product *Product) Print() {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("ID: %d\n", product.ID))
	builder.WriteString(fmt.Sprintf("Serial: %s\n", product.Serial))
	builder.WriteString(fmt.Sprintf("Name: %s\n", product.Name))
	builder.WriteString(fmt.Sprintf("Description: %s\n", product.Description))
	builder.WriteString(fmt.Sprintf("Brand: %s\n", product.Brand))
	builder.WriteString(fmt.Sprintf("Product Status: %s\n", &product.Status))
	builder.WriteString(fmt.Sprintf("Category: %s\n", product.Category))
	builder.WriteString(fmt.Sprintf("Subcategory: %s\n", product.Subcategory))
	builder.WriteString(fmt.Sprintf("Colours: %s\n", product.Colours))
	builder.WriteString(fmt.Sprintf("Sizes: %s\n", product.Sizes))
	builder.WriteString(fmt.Sprintf("Segmentation: %s\n", product.Segmentation))
	builder.WriteString(fmt.Sprintf("Unit Price w/o VAT: %s\n", product.UnitPriceWithoutVat.Display()))
	builder.WriteString(fmt.Sprintf("Unit Price w VAT: %s\n", product.UnitPriceWithVat.Display()))
	builder.WriteString(fmt.Sprintf("Created At %s\n", product.CreatedAt))
	builder.WriteString(fmt.Sprintf("Updated At %s\n", product.UpdatedAt))
	builder.WriteString(fmt.Sprintf("Deleted At %s\n", product.DeletedAt))
	fmt.Println(builder.String())
}

func (product *Product) Duplicate() *Product {
	newProduct := Product{
		ID:                  product.ID,
		Serial:              product.Serial,
		Name:                product.Name,
		Description:         product.Description,
		Brand:               product.Brand,
		Status:              product.Status,
		Category:            product.Category,
		Subcategory:         product.Subcategory,
		Colours:             product.Colours,
		Sizes:               product.Sizes,
		Segmentation:        product.Segmentation,
		UnitPriceWithoutVat: product.UnitPriceWithoutVat,
		UnitPriceWithVat:    product.UnitPriceWithVat,
		CreatedAt:           product.CreatedAt,
		UpdatedAt:           product.UpdatedAt,
		DeletedAt:           product.DeletedAt,
	}
	return &newProduct
}

func (product *Product) GetDBID(appCtx *internal.AppContext) int64 {
	ctx := context.Background()
	existingProductID, err := appCtx.Queries.GetProductIDBySerial(ctx, product.Serial)
	if err != nil {
		return 0
	}
	return existingProductID
}

func (product *Product) GetCategoryID(appCtx *internal.AppContext) (int64, error) {
	ctx := context.Background()
	var categoryID int64
	if product.Category != "" {
		existingProductCategory, err := appCtx.Queries.GetProductCategoryByCategoryAndSubcategory(
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
			existingProductCategory, err = appCtx.Queries.CreateProductCategory(
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
	return categoryID, nil
}

func (product *Product) InsertToDB(appCtx *internal.AppContext) (int64, error) {
	ctx := context.Background()
	categoryID, err := product.GetCategoryID(appCtx)
	now := time.Now().UTC()
	insertedProduct, err := appCtx.Queries.CreateProduct(
		ctx,
		cchoice_db.CreateProductParams{
			Serial: product.Serial,
			Name:   product.Name,
			Description: sql.NullString{
				String: product.Description,
				Valid:  true,
			},
			Brand:  product.Brand,
			Status: product.Status.String(),
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
			DeletedAt: constants.DT_BEGINNING,
		},
	)
	if err != nil {
		return 0, err
	}

	return insertedProduct.ID, nil
}

func (product *Product) UpdateToDB(appCtx *internal.AppContext) (int64, error) {
	ctx := context.Background()
	categoryID, err := product.GetCategoryID(appCtx)
	now := time.Now().UTC()

	updatedID, err := appCtx.Queries.UpdateProduct(
		ctx,
		cchoice_db.UpdateProductParams{
			ID:   product.ID,
			Name: product.Name,
			Description: sql.NullString{
				String: product.Description,
				Valid:  true,
			},
			Brand:  product.Brand,
			Status: product.Status.String(),
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
			UpdatedAt:                   now,
		},
	)
	return updatedID, err
}

func DBRowToProduct(row *cchoice_db.GetProductByIDRow) *Product {
	dbp := &Product{
		ID:           row.ID,
		Serial:       row.Serial,
		Name:         row.Name,
		Description:  row.Description.String,
		Brand:        row.Brand,
		Status:       ParseProductStatusEnum(row.Status),
		Category:     row.Category.String,
		Subcategory:  row.Subcategory.String,
		Colours:      row.Colours.String,
		Sizes:        row.Sizes.String,
		Segmentation: row.Segmentation.String,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		DeletedAt:    row.DeletedAt,
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

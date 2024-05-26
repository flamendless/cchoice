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

type ProductCategory struct {
	ID          int64
	Category    string
	Subcategory string
}

type ProductSpecs struct {
	ID            int64
	Colours       string
	Sizes         string
	Segmentation  string
	PartNumber    string
	Power         string
	Capacity      string
	ScopeOfSupply string
}

type Product struct {
	ID                  int64
	Serial              string
	Name                string
	Description         string
	Brand               string
	Status              ProductStatus
	ProductCategory     *ProductCategory
	ProductSpecs        *ProductSpecs
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
	product.ProductCategory.Category = slug.Make(product.ProductCategory.Category)
	product.ProductCategory.Subcategory = slug.Make(product.ProductCategory.Subcategory)
}

func (product *Product) Print() {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("ID: %d\n", product.ID))
	builder.WriteString(fmt.Sprintf("Serial: %s\n", product.Serial))
	builder.WriteString(fmt.Sprintf("Name: %s\n", product.Name))
	builder.WriteString(fmt.Sprintf("Description: %s\n", product.Description))
	builder.WriteString(fmt.Sprintf("Brand: %s\n", product.Brand))
	builder.WriteString(fmt.Sprintf("Product Status: %s\n", &product.Status))
	product.ProductCategory.Print(&builder)
	product.ProductSpecs.Print(&builder)
	builder.WriteString(fmt.Sprintf("Unit Price w/o VAT: %s\n", product.UnitPriceWithoutVat.Display()))
	builder.WriteString(fmt.Sprintf("Unit Price w VAT: %s\n", product.UnitPriceWithVat.Display()))
	builder.WriteString(fmt.Sprintf("Created At %s\n", product.CreatedAt))
	builder.WriteString(fmt.Sprintf("Updated At %s\n", product.UpdatedAt))
	builder.WriteString(fmt.Sprintf("Deleted At %s\n", product.DeletedAt))
	fmt.Println(builder.String())
}

func (productSpecs *ProductSpecs) Print(builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf("Colours: %s\n", productSpecs.Colours))
	builder.WriteString(fmt.Sprintf("Sizes: %s\n", productSpecs.Sizes))
	builder.WriteString(fmt.Sprintf("Segmentation: %s\n", productSpecs.Segmentation))
	builder.WriteString(fmt.Sprintf("Part Number: %s\n", productSpecs.PartNumber))
	builder.WriteString(fmt.Sprintf("Power: %s\n", productSpecs.Power))
	builder.WriteString(fmt.Sprintf("Capacity: %s\n", productSpecs.Capacity))
	builder.WriteString(fmt.Sprintf("Scope of Supply: %s\n", productSpecs.ScopeOfSupply))
}

func (productCategory *ProductCategory) Print(builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf("Category: %s\n", productCategory.Category))
	builder.WriteString(fmt.Sprintf("Subcategory: %s\n", productCategory.Subcategory))
}

func (product *Product) Duplicate() *Product {
	newProduct := Product{
		ID:                  product.ID,
		Serial:              product.Serial,
		Name:                product.Name,
		Description:         product.Description,
		Brand:               product.Brand,
		Status:              product.Status,
		ProductCategory:     product.ProductCategory,
		ProductSpecs:        product.ProductSpecs,
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
	existingProductID, err := appCtx.QueriesRead.GetProductIDBySerial(ctx, product.Serial)
	if err != nil {
		return 0
	}
	return existingProductID
}

func (product *Product) GetOrInsertCategoryID(appCtx *internal.AppContext) (int64, error) {
	ctx := context.Background()
	var categoryID int64
	if product.ProductCategory.ID != 0 {
		existingProductCategory, err := appCtx.QueriesRead.GetProductCategoryByCategoryAndSubcategory(
			ctx,
			cchoice_db.GetProductCategoryByCategoryAndSubcategoryParams{
				Category: sql.NullString{
					String: product.ProductCategory.Category,
					Valid:  true,
				},
				Subcategory: sql.NullString{
					String: product.ProductCategory.Subcategory,
					Valid:  true,
				},
			},
		)
		if err != nil {
			existingProductCategory, err = appCtx.Queries.CreateProductCategory(
				ctx,
				cchoice_db.CreateProductCategoryParams{
					Category: sql.NullString{
						String: product.ProductCategory.Category,
						Valid:  true,
					},
					Subcategory: sql.NullString{
						String: product.ProductCategory.Subcategory,
						Valid:  product.ProductCategory.Subcategory != "",
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

func (productSpecs *ProductSpecs) GetOrInsertProductSpecsID(appCtx *internal.AppContext) (int64, error) {
	// ctx := context.Background()
	var productSpecsID int64
	return productSpecsID, nil
}

func (product *Product) InsertToDB(appCtx *internal.AppContext) (int64, error) {
	ctx := context.Background()
	categoryID, err := product.GetOrInsertCategoryID(appCtx)
	if err != nil {
		return 0, err
	}

	productSpecsID, err := product.ProductSpecs.GetOrInsertProductSpecsID(appCtx)
	if err != nil {
		return 0, err
	}

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
			ProductSpecsID: sql.NullInt64{
				Int64: productSpecsID,
				Valid: productSpecsID != 0,
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
	categoryID, err := product.GetOrInsertCategoryID(appCtx)
	productSpecsID, err := product.ProductSpecs.GetOrInsertProductSpecsID(appCtx)
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
			ProductSpecsID: sql.NullInt64{
				Int64: productSpecsID,
				Valid: productSpecsID != 0,
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
		ID:          row.ID,
		Serial:      row.Serial,
		Name:        row.Name,
		Description: row.Description.String,
		Brand:       row.Brand,
		Status:      ParseProductStatusEnum(row.Status),
		ProductCategory: &ProductCategory{
			Category:    row.Category.String,
			Subcategory: row.Subcategory.String,
		},
		ProductSpecs: &ProductSpecs{
			Colours:       row.Colours.String,
			Sizes:         row.Sizes.String,
			Segmentation:  row.Segmentation.String,
			PartNumber:    row.PartNumber.String,
			Power:         row.Power.String,
			Capacity:      row.Capacity.String,
			ScopeOfSupply: row.ScopeOfSupply.String,
		},
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: row.DeletedAt,
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

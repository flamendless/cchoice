package models

import (
	cchoice_db "cchoice/cchoice_db"
	"cchoice/internal/constants"
	"cchoice/internal/ctx"
	"cchoice/internal/enums"
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
	ProductID   int64
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
	Status              enums.ProductStatus
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
	product.Serial = fmt.Sprintf("%s-%s-%d-%d", brandInitials, nameInitials, rowIdx, product.ID)
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

func (product *Product) GetDBID(ctxDB *ctx.Database) int64 {
	ctx := context.Background()
	existingProductID, err := ctxDB.QueriesRead.GetProductIDBySerial(ctx, product.Serial)
	if err != nil {
		return 0
	}
	product.ID = existingProductID
	return existingProductID
}

func (product *Product) GetOrInsertCategoryID(ctxDB *ctx.Database) (int64, error) {
	ctx := context.Background()
	var categoryID int64
	if product.ProductCategory.ID != 0 {
		return 0, nil
	}

	existingProductCategory, err := ctxDB.QueriesRead.GetProductCategoryByCategoryAndSubcategory(
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
	if err == nil {
		categoryID = existingProductCategory.ID
	} else {
		newProductCategory, err := ctxDB.Queries.CreateProductCategory(
			ctx,
			cchoice_db.CreateProductCategoryParams{
				ProductID: product.ID,
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
		categoryID = newProductCategory.ID
	}
	return categoryID, nil
}

func (product *Product) GetOrInsertProductSpecsID(ctxDB *ctx.Database) (int64, error) {
	ctx := context.Background()
	var productSpecsID int64

	if product.ProductSpecs.ID != 0 {
		return productSpecsID, nil
	}

	existingProductSpecs, err := ctxDB.QueriesRead.GetProductSpecsByProductID(ctx, product.ID)
	if err == nil {
		productSpecsID = existingProductSpecs.ID
	} else {
		newProductSpecs, err := ctxDB.Queries.CreateProductSpecs(
			ctx,
			cchoice_db.CreateProductSpecsParams{
				Colours: sql.NullString{
					String: product.ProductSpecs.Colours,
					Valid:  true,
				},
				Sizes: sql.NullString{
					String: product.ProductSpecs.Sizes,
					Valid:  true,
				},
				Segmentation: sql.NullString{
					String: product.ProductSpecs.Segmentation,
					Valid:  true,
				},
				PartNumber: sql.NullString{
					String: product.ProductSpecs.PartNumber,
					Valid:  true,
				},
				Power: sql.NullString{
					String: product.ProductSpecs.Power,
					Valid:  true,
				},
				Capacity: sql.NullString{
					String: product.ProductSpecs.Capacity,
					Valid:  true,
				},
				ScopeOfSupply: sql.NullString{
					String: product.ProductSpecs.ScopeOfSupply,
					Valid:  true,
				},
			},
		)
		if err != nil {
			logs.Log().Warn(
				"Insert product specs",
				zap.Error(err),
			)
		}
		productSpecsID = newProductSpecs.ID
	}
	return productSpecsID, nil
}

func (product *Product) InsertToDB(ctxDB *ctx.Database) (int64, error) {
	ctx := context.Background()

	productSpecsID, err := product.GetOrInsertProductSpecsID(ctxDB)
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	insertedProduct, err := ctxDB.Queries.CreateProduct(
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

	product.ID = insertedProduct.ID

	_, err = product.GetOrInsertCategoryID(ctxDB)
	if err != nil {
		return 0, err
	}

	return insertedProduct.ID, nil
}

func (product *Product) UpdateToDB(ctxDB *ctx.Database) (int64, error) {
	ctx := context.Background()
	productSpecsID, err := product.GetOrInsertProductSpecsID(ctxDB)

	now := time.Now().UTC()
	updatedID, err := ctxDB.Queries.UpdateProduct(
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

	product.ID = updatedID
	_, err = product.GetOrInsertCategoryID(ctxDB)

	return updatedID, err
}

func DBRowToProduct(row *cchoice_db.GetProductBySerialRow) *Product {
	unitPriceWithoutVat := decimal.NewFromInt(row.UnitPriceWithoutVat / 100)
	unitPriceWithVat := decimal.NewFromInt(row.UnitPriceWithVat / 100)

	return &Product{
		ID:          row.ID,
		Serial:      row.Serial,
		Name:        row.Name,
		Description: row.Description.String,
		Brand:       row.Brand,
		Status:      enums.ParseProductStatusEnum(row.Status),
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
		UnitPriceWithoutVat: money.New(
			unitPriceWithoutVat.CoefficientInt64(),
			row.UnitPriceWithoutVatCurrency,
		),
		UnitPriceWithVat: money.New(
			unitPriceWithVat.CoefficientInt64(),
			row.UnitPriceWithVatCurrency,
		),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: row.DeletedAt,
	}
}

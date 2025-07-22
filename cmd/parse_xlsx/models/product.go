package models

import (
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/gosimple/slug"
	"go.uber.org/zap"
)

type ProductCategory struct {
	Category    string
	Subcategory string
	ID          int64
}

type ProductSpecs struct {
	Colours       string
	Sizes         string
	Segmentation  string
	PartNumber    string
	Power         string
	Capacity      string
	ScopeOfSupply string
	ID            int64
}

type Product struct {
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           time.Time
	Brand               *Brand
	ProductCategory     *ProductCategory
	ProductSpecs        *ProductSpecs
	UnitPriceWithoutVat *money.Money
	UnitPriceWithVat    *money.Money
	Serial              string
	Name                string
	Description         string
	ID                  int64
	Status              enums.ProductStatus
}

func (product *Product) PostProcess(rowIdx int) {
	if product == nil {
		panic("nil product")
	}
	brandInitials := utils.GetInitials(product.Brand.Name)
	nameInitials := utils.GetInitials(product.Name)
	product.Serial = fmt.Sprintf("%s-%s-%d-%d", brandInitials, nameInitials, rowIdx, product.ID)
	if product.ProductCategory != nil {
		product.ProductCategory.Category = slug.Make(product.ProductCategory.Category)
		product.ProductCategory.Subcategory = slug.Make(product.ProductCategory.Subcategory)
	}
}

func (product *Product) Print() {
	builder := strings.Builder{}
	fmt.Fprintf(&builder, "ID: %d\n", product.ID)
	fmt.Fprintf(&builder, "Serial: %s\n", product.Serial)
	fmt.Fprintf(&builder, "Name: %s\n", product.Name)
	fmt.Fprintf(&builder, "Description: %s\n", product.Description)
	fmt.Fprintf(&builder, "Brand: %s\n", product.Brand.Name)
	fmt.Fprintf(&builder, "Product Status: %s\n", &product.Status)
	product.ProductCategory.Print(&builder)
	product.ProductSpecs.Print(&builder)
	fmt.Fprintf(&builder, "Unit Price w/o VAT: %s\n", product.UnitPriceWithoutVat.Display())
	fmt.Fprintf(&builder, "Unit Price w VAT: %s\n", product.UnitPriceWithVat.Display())
	fmt.Fprintf(&builder, "Created At %s\n", product.CreatedAt)
	fmt.Fprintf(&builder, "Updated At %s\n", product.UpdatedAt)
	fmt.Fprintf(&builder, "Deleted At %s\n", product.DeletedAt)
	fmt.Println(builder.String())
}

func (productSpecs *ProductSpecs) Print(builder *strings.Builder) {
	fmt.Fprintf(builder, "Colours: %s\n", productSpecs.Colours)
	fmt.Fprintf(builder, "Sizes: %s\n", productSpecs.Sizes)
	fmt.Fprintf(builder, "Segmentation: %s\n", productSpecs.Segmentation)
	fmt.Fprintf(builder, "Part Number: %s\n", productSpecs.PartNumber)
	fmt.Fprintf(builder, "Power: %s\n", productSpecs.Power)
	fmt.Fprintf(builder, "Capacity: %s\n", productSpecs.Capacity)
	fmt.Fprintf(builder, "Scope of Supply: %s\n", productSpecs.ScopeOfSupply)
}

func (productCategory *ProductCategory) Print(builder *strings.Builder) {
	fmt.Fprintf(builder, "Category: %s\n", productCategory.Category)
	fmt.Fprintf(builder, "Subcategory: %s\n", productCategory.Subcategory)
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

func (product *Product) GetDBID(ctx context.Context, db database.Service) int64 {
	existingProductID, err := db.GetQueries().GetProductIDBySerial(ctx, product.Serial)
	if err != nil {
		return 0
	}
	product.ID = existingProductID
	return existingProductID
}

func (product *Product) GetOrInsertCategoryID(ctx context.Context, db database.Service) (int64, error) {
	var categoryID int64

	existingProductCategory, err := db.GetQueries().GetProductCategoryByCategoryAndSubcategory(
		ctx,
		queries.GetProductCategoryByCategoryAndSubcategoryParams{
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
	} else if errors.Is(err, sql.ErrNoRows) {
		newProductCategory, err := db.GetQueries().CreateProductCategory(
			ctx,
			queries.CreateProductCategoryParams{
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
			logs.Log().Warn(
				"Insert product category",
				zap.Error(err),
			)
			return 0, err
		}
		categoryID = newProductCategory.ID
	}

	_, err = db.GetQueries().GetProductsCategoriesByIDs(
		ctx,
		queries.GetProductsCategoriesByIDsParams{
			CategoryID: categoryID,
			ProductID:  product.ID,
		},
	)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		_, err = db.GetQueries().CreateProductsCategories(
			ctx,
			queries.CreateProductsCategoriesParams{
				CategoryID: categoryID,
				ProductID:  product.ID,
			},
		)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logs.Log().Warn(
				"Insert products categories",
				zap.Error(err),
			)
			return 0, err
		}
	}

	return categoryID, nil
}

func (product *Product) GetOrInsertProductSpecsID(ctx context.Context, db database.Service) (int64, error) {
	var productSpecsID int64

	existingProductSpecs, err := db.GetQueries().GetProductSpecsByProductID(ctx, product.ID)
	if err == nil {
		productSpecsID = existingProductSpecs.ID
	} else {
		newProductSpecs, err := db.GetQueries().CreateProductSpecs(
			ctx,
			queries.CreateProductSpecsParams{
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

func (product *Product) InsertToDB(ctx context.Context, db database.Service) (int64, error) {
	productSpecsID, err := product.GetOrInsertProductSpecsID(ctx, db)
	if err != nil {
		return 0, err
	}
	product.ProductSpecs.ID = productSpecsID

	now := time.Now().UTC()
	insertedProduct, err := db.GetQueries().CreateProducts(
		ctx,
		queries.CreateProductsParams{
			Serial: product.Serial,
			Name:   product.Name,
			Description: sql.NullString{
				String: product.Description,
				Valid:  true,
			},
			BrandID: product.Brand.ID,
			Status:  product.Status.String(),
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
			DeletedAt: constants.DtBeginning,
		},
	)
	if err != nil {
		return 0, err
	}

	product.ID = insertedProduct.ID

	categoryID, err := product.GetOrInsertCategoryID(ctx, db)
	if err != nil {
		return 0, err
	}
	product.ProductCategory.ID = categoryID

	return insertedProduct.ID, nil
}

func (product *Product) UpdateToDB(ctx context.Context, db database.Service) (int64, error) {
	productSpecsID, err := product.GetOrInsertProductSpecsID(ctx, db)
	if err != nil {
		return 0, err
	}
	product.ProductSpecs.ID = productSpecsID

	now := time.Now().UTC()
	updatedID, err := db.GetQueries().UpdateProducts(
		ctx,
		queries.UpdateProductsParams{
			ID:   product.ID,
			Name: product.Name,
			Description: sql.NullString{
				String: product.Description,
				Valid:  true,
			},
			BrandID: product.Brand.ID,
			Status:  product.Status.String(),
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
	if err != nil {
		return 0, err
	}

	product.ID = updatedID
	categoryID, err := product.GetOrInsertCategoryID(ctx, db)
	if err != nil {
		return 0, err
	}
	product.ProductCategory.ID = categoryID

	return updatedID, err
}

func DBRowToProduct(row *queries.GetProductsBySerialRow) *Product {
	moneyWithoutVat := utils.NewMoney(row.UnitPriceWithoutVat, row.UnitPriceWithoutVatCurrency)
	moneyWithVat := utils.NewMoney(row.UnitPriceWithVat, row.UnitPriceWithVatCurrency)

	return &Product{
		ID:          row.ID,
		Serial:      row.Serial,
		Name:        row.Name,
		Description: row.Description.String,
		Brand:       &Brand{},
		Status:      enums.ParseProductStatusToEnum(row.Status),
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
		UnitPriceWithoutVat: moneyWithoutVat,
		UnitPriceWithVat:    moneyWithVat,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
		DeletedAt:           row.DeletedAt,
	}
}

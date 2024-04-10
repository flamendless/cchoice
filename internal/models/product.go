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

package models

import (
	"fmt"
	"strings"

	"github.com/Rhymond/go-money"
)

type Product struct {
	Serial              string
	Name                string
	Description         string
	Category            string
	Subcategory         string
	Colours             string
	Sizes               string
	Segmentation        string
	UnitPriceWithoutVat *money.Money
	UnitPriceWithVat    *money.Money
}

func (product *Product) Finalize() {
}

func (product *Product) Print() {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Serial: %s\n", product.Serial))
	builder.WriteString(fmt.Sprintf("Name: %s\n", product.Name))
	builder.WriteString(fmt.Sprintf("Description: %s\n", product.Description))

	builder.WriteString(fmt.Sprintf("Category: %s\n", product.Category))
	builder.WriteString(fmt.Sprintf("Subcategory: %s\n", product.Subcategory))

	builder.WriteString(fmt.Sprintf("Colours: %s\n", product.Colours))
	builder.WriteString(fmt.Sprintf("Sizes: %s\n", product.Sizes))
	builder.WriteString(fmt.Sprintf("Segmentation: %s\n", product.Segmentation))

	builder.WriteString(fmt.Sprintf("Unit Price w/o VAT: %s\n", product.UnitPriceWithoutVat.Display()))
	builder.WriteString(fmt.Sprintf("Unit Price w VAT: %s\n", product.UnitPriceWithVat.Display()))

	fmt.Println(builder.String())
}

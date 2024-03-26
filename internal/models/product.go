package models

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type Product struct {
	Serial      string
	Name        string
	Description string
	UnitPrice   decimal.Decimal
}

func (product *Product) Print() {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Serial: %s\n", product.Serial))
	builder.WriteString(fmt.Sprintf("Name: %s\n", product.Name))
	builder.WriteString(fmt.Sprintf("Description: %s\n", product.Description))
	builder.WriteString(fmt.Sprintf("Unit Price: %s\n", product.UnitPrice))
	fmt.Println(builder.String())
}

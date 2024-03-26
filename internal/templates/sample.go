package templates

import (
	"cchoice/internal/models"
	"fmt"

	"github.com/shopspring/decimal"
)

var SampleColumns map[string]*Column = map[string]*Column{
	"Product Name": {
		Index:    -1,
		Required: true,
	},
	"Product Number": {
		Index:    -1,
		Required: true,
	},
	"Description": {
		Index:    -1,
		Required: true,
	},
	"Unit Price": {
		Index:    -1,
		Required: true,
	},
}

func SampleRowToProduct(tpl *Template, row []string) (*models.Product, error) {
	idxSerial := tpl.Columns["Product Number"].Index
	idxProductName := tpl.Columns["Product Name"].Index
	idxDesc := tpl.Columns["Description"].Index
	idxUnitPrice := tpl.Columns["Unit Price"].Index

	unitPrice, err := decimal.NewFromString(row[idxUnitPrice])
	if err != nil {
		fmt.Printf("%s - '%s'\n", err, row[idxUnitPrice])
		unitPrice = decimal.Zero
	}

	return &models.Product{
		Serial:      row[idxSerial],
		Name:        row[idxProductName],
		Description: row[idxDesc],
		UnitPrice:   unitPrice,
	}, nil
}

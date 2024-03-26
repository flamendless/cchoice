package templates

import (
	"cchoice/internal/models"

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
	colSerial := tpl.Columns["Product Number"].Index
	colProductName := tpl.Columns["Product Name"].Index
	colDesc := tpl.Columns["Description"].Index
	colUnitPrice := tpl.Columns["Unit Price"].Index

	unitPrice, err := decimal.NewFromString(row[colUnitPrice])
	if err != nil {
		return nil, err
	}

	return &models.Product{
		Serial:      row[colSerial],
		Name:        row[colProductName],
		Description: row[colDesc],
		UnitPrice:   unitPrice,
	}, nil
}

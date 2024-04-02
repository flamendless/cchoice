package templates

import (
	"cchoice/internal/models"
	"errors"

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

	var errs error

	unitPrice, err := decimal.NewFromString(row[idxUnitPrice])
	if err != nil {
		errs = errors.Join(errs, err)
	}

	name := row[idxProductName]
	if name == "" {
		errs = errors.Join(errs, errors.New("'Product Name' must not be empty"))
	}

	if errs != nil {
		return nil, errs
	}

	return &models.Product{
		Serial:      row[idxSerial],
		Name:        name,
		Description: row[idxDesc],
		UnitPrice:   unitPrice,
	}, nil
}

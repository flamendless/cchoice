package templates

import (
	"cchoice/internal/models"
	"cchoice/internal/utils"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
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

	price := row[idxUnitPrice]
	errPrice := utils.ValidateNotBlank(price, "unit price")
	if errPrice != nil {
		errs = errors.Join(errs, errPrice)
	}
	unitPrice, err := decimal.NewFromString(price)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	name := row[idxProductName]
	errProductName := utils.ValidateNotBlank(name, "product name")
	if errProductName != nil {
		errs = errors.Join(errs, errProductName)
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


func SampleProcessRows(tpl *Template, rows *excelize.Rows) []*models.Product {
	var products []*models.Product = make([]*models.Product, 0, tpl.AssumedRowsCount)

	rows.Next()

	rowIdx := 0

	for rows.Next() {
		rowIdx++
		row, err := rows.Columns()

		if err != nil {
			fmt.Println(err)
			return products
		}

		if len(row) == 0 {
			break
		}

		row = tpl.AlignRow(row)
		product, errs := tpl.RowToProduct(tpl, row)
		if errs != nil {
			if tpl.AppContext.Strict {
				fmt.Println(errs)
				panic("error immediately")
			}
			fmt.Printf("row %d: %s\n", rowIdx, errs)
			continue
		}

		if (tpl.AppContext.Limit != 0) && (rowIdx > tpl.AppContext.Limit) {
			return products
		}

		products = append(products, product)
	}

	return products
}

package templates

import (
	"cchoice/internal/logs"
	"cchoice/internal/models"
	"cchoice/internal/utils"
	"errors"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
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

	serial := row[idxSerial]
	name := row[idxProductName]
	desc := row[idxDesc]
	price := row[idxUnitPrice]

	var errs error

	unitPriceMoney, err := utils.SanitizePrice(price)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	errProductName := utils.ValidateNotBlank(name, "product name")
	if errProductName != nil {
		errs = errors.Join(errs, errProductName)
	}

	if errs != nil {
		return nil, errs
	}

	return &models.Product{
		Serial:              serial,
		Name:                name,
		Description:         desc,
		UnitPriceWithoutVat: unitPriceMoney,
		UnitPriceWithVat:    unitPriceMoney,
	}, nil
}

func SampleProcessRows(tpl *Template, rows *excelize.Rows) []*models.Product {
	var products []*models.Product = make([]*models.Product, 0, tpl.AssumedRowsCount)

	rowIdx := 0
	for i := 0; i < tpl.SkipInitialRows+1; i++ {
		rows.Next()
		rowIdx++
	}

	for rows.Next() {
		rowIdx++
		row, err := rows.Columns()

		if err != nil {
			logs.Log().Error(err.Error())
			return products
		}

		if len(row) == 0 {
			break
		}

		row = tpl.AlignRow(row)
		product, errs := tpl.RowToProduct(tpl, row)
		if errs != nil {
			if tpl.AppContext.Strict {
				logs.Log().Panic(errs.Error())
				return nil
			}
			logs.Log().Info(
				"processed row to product",
				zap.Int("row", rowIdx),
				zap.String("errors", errs.Error()),
			)
			continue
		}

		if (tpl.AppContext.Limit != 0) && (rowIdx > tpl.AppContext.Limit) {
			return products
		}

		products = append(products, product)
	}

	return products
}

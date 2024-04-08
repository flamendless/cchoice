package templates

import (
	"cchoice/internal/logs"
	"cchoice/internal/models"
	"cchoice/internal/utils"

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

func SampleRowToProduct(tpl *Template, row []string) (*models.Product, []error) {
	idxSerial := tpl.Columns["Product Number"].Index
	idxProductName := tpl.Columns["Product Name"].Index
	idxDesc := tpl.Columns["Description"].Index
	idxUnitPrice := tpl.Columns["Unit Price"].Index

	serial := row[idxSerial]
	name := row[idxProductName]
	desc := row[idxDesc]
	price := row[idxUnitPrice]

	errs := make([]error, 0, 4)

	unitPriceMoney, err := utils.SanitizePrice(price)
	if len(err) > 0 {
		errs = append(errs, err...)
	}

	errProductName := utils.ValidateNotBlank(name, "product name")
	if errProductName != nil {
		errs = append(errs, err...)
	}

	if len(errs) > 0 {
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

	totalErrors := 0

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
		if len(errs) > 0 {
			if tpl.AppContext.Strict {
				logs.Log().Panic("error", zap.Errors("errors", errs))
				return nil
			}
			logs.Log().Info(
				"row to product",
				zap.Int("row number", rowIdx),
				zap.Errors("errors", errs),
			)
			totalErrors += 1
			continue
		}

		product.PostProcess()

		if (tpl.AppContext.Limit > 0) && (rowIdx > tpl.AppContext.Limit) {
			return products
		}

		products = append(products, product)
	}

	logs.Log().Info(
		"results",
		zap.Int("processed", len(products)),
		zap.Int("errors", totalErrors),
		zap.Int("total", len(products) + totalErrors),
	)

	return products
}

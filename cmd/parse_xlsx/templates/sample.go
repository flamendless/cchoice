package templates

import (
	"cchoice/cmd/parse_xlsx/models"
	"cchoice/internal/logs"
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
	colProductNumber, ok := tpl.Columns["Product Number"]
	if !ok {
		panic("missing column")
	}
	colProductName, ok := tpl.Columns["Product Name"]
	if !ok {
		panic("missing column")
	}
	colDescription, ok := tpl.Columns["Description"]
	if !ok {
		panic("missing column")
	}
	colUnitPrice, ok := tpl.Columns["Unit Price"]
	if !ok {
		panic("missing column")
	}

	serial := row[colProductNumber.Index]
	name := row[colProductName.Index]
	desc := row[colDescription.Index]
	price := row[colUnitPrice.Index]

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
		Brand:               tpl.Brand,
		Description:         desc,
		UnitPriceWithoutVat: unitPriceMoney,
		UnitPriceWithVat:    unitPriceMoney,
	}, nil
}

func SampleRowToSpecs(tpl *Template, row []string) *models.ProductSpecs {
	return &models.ProductSpecs{}
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
			if tpl.AppFlags.Strict {
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

		product.PostProcess(rowIdx)

		if (tpl.AppFlags.Limit > 0) && (rowIdx > tpl.AppFlags.Limit) {
			return products
		}

		products = append(products, product)
	}

	logs.Log().Info(
		"results",
		zap.Int("processed", len(products)),
		zap.Int("errors", totalErrors),
		zap.Int("total", len(products)+totalErrors),
	)

	return products
}

func SampleGetPromotedCategories() []string {
	return nil
}

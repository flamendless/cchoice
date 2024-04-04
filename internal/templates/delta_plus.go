package templates

import (
	"cchoice/internal/logs"
	"cchoice/internal/models"
	"cchoice/internal/utils"
	"errors"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

var DeltaPlusColumns map[string]*Column = map[string]*Column{
	"ARTICLE": {
		Index:    -1,
		Required: true,
	},
	"COLOURS": {
		Index:    -1,
		Required: true,
	},
	"SIZES": {
		Index:    -1,
		Required: true,
	},
	"SEGMENTATION": {
		Index:    -1,
		Required: true,
	},
	"DESCRIPTION": {
		Index:    -1,
		Required: true,
	},
	"END USER UNIT PRICE PHP WITHOUT VAT*": {
		Index:    -1,
		Required: true,
	},
	"END USER UNIT PRICE PHP WITH VAT* (SRP)": {
		Index:    -1,
		Required: true,
	},
}

func DeltaPlusRowToProduct(tpl *Template, row []string) (*models.Product, error) {
	idxArticle := tpl.Columns["ARTICLE"].Index
	idxColours := tpl.Columns["COLOURS"].Index
	idxSizes := tpl.Columns["SIZES"].Index
	idxSegmentation := tpl.Columns["SEGMENTATION"].Index
	idxDesc := tpl.Columns["DESCRIPTION"].Index
	idxPriceWithoutVat := tpl.Columns["END USER UNIT PRICE PHP WITHOUT VAT*"].Index
	idxPriceWithVat := tpl.Columns["END USER UNIT PRICE PHP WITH VAT* (SRP)"].Index

	name := row[idxArticle]
	colours := row[idxColours]
	sizes := utils.SanitizeSize(row[idxSizes])
	segmentation := row[idxSegmentation]
	desc := row[idxDesc]
	priceWithoutVat := row[idxPriceWithoutVat]
	priceWithVat := row[idxPriceWithVat]

	var errs error

	errProductName := utils.ValidateNotBlank(name, "article")
	if errProductName != nil {
		errs = errors.Join(errs, errProductName)
	}

	unitPriceWithoutVat, err := utils.SanitizePrice(priceWithoutVat)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	unitPriceWithVat, err := utils.SanitizePrice(priceWithVat)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return nil, errs
	}

	return &models.Product{
		Name:                name,
		Description:         desc,
		Colours:             colours,
		Sizes:               sizes,
		Segmentation:        segmentation,
		UnitPriceWithoutVat: unitPriceWithoutVat,
		UnitPriceWithVat:    unitPriceWithVat,
	}, nil
}

func DeltaPlusGetCategory(t *Template, row []string) string {
	if len(row) != 1 {
		return ""
	}
	return row[0]
}

func DeltaPlusProcessRows(tpl *Template, rows *excelize.Rows) []*models.Product {
	var products []*models.Product = make([]*models.Product, 0, tpl.AssumedRowsCount)

	rowIdx := 0
	for i := 0; i < tpl.SkipInitialRows+1; i++ {
		rows.Next()
		rowIdx++
	}

	var previousRow []string
	category := ""
	subcategory := ""

	for rows.Next() {
		rowIdx++
		row, err := rows.Columns()
		if err != nil {
			logs.Log().Error(err.Error())
			return products
		}

		if len(row) == 0 {
			break
		} else if len(row) == 1 {
			if (category == "") && (subcategory == "") {
				category = row[0]
			} else if (category != "") && (subcategory == "") {
				subcategory = row[0]
			} else if (category != "") && (subcategory != "") {
				if len(previousRow) != 1 {
					subcategory = row[0]
				} else {
					category = previousRow[0]
					subcategory = row[0]
				}
			}
			previousRow = row
			continue
		}

		row = tpl.AlignRow(row)
		product, errs := tpl.RowToProduct(tpl, row)
		if errs != nil {
			if tpl.AppContext.Strict {
				logs.Log().Error(errs.Error())
				return nil
			}
			logs.Log().Info(
				"processed row to product",
				zap.Int("row", rowIdx),
				zap.String("errors", errs.Error()),
			)
			previousRow = row
			continue
		}

		product.Category = category
		product.Subcategory = subcategory

		if (tpl.AppContext.Limit != 0) && (rowIdx > tpl.AppContext.Limit) {
			return products
		}

		products = append(products, product)
		previousRow = row
	}

	return products
}

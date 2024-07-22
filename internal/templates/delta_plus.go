package templates

import (
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/models"
	"cchoice/internal/utils"
	"strings"

	"github.com/Rhymond/go-money"
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

func DeltaPlusProcessPrices(tpl *Template, row []string) ([]*money.Money, []error) {
	idxPriceWithoutVat := tpl.Columns["END USER UNIT PRICE PHP WITHOUT VAT*"].Index
	idxPriceWithVat := tpl.Columns["END USER UNIT PRICE PHP WITH VAT* (SRP)"].Index
	priceWithoutVat := row[idxPriceWithoutVat]
	priceWithVat := row[idxPriceWithVat]

	prices := make([]*money.Money, 0, 8)
	errs := make([]error, 0, 8)

	unitPriceWithoutVat, err := utils.SanitizePrice(priceWithoutVat)
	if err != nil {
		errs = append(errs, err...)
	}

	unitPriceWithVat, err := utils.SanitizePrice(priceWithVat)
	if err != nil {
		errs = append(errs, err...)
	}

	if len(errs) == 0 {
		prices = append(prices, unitPriceWithoutVat)
		prices = append(prices, unitPriceWithVat)
	}

	return prices, errs
}

func DeltaPlusRowToProduct(tpl *Template, row []string) (*models.Product, []error) {
	errsRes := make([]error, 0, 8)

	idxArticle := tpl.Columns["ARTICLE"].Index
	idxDesc := tpl.Columns["DESCRIPTION"].Index
	name := row[idxArticle]

	var status enums.ProductStatus
	if strings.Contains(strings.ToLower(name), "discontinued") {
		status = enums.PRODUCT_STATUS_DELETED
		parserErr := errs.NewParserError(errs.ProductDiscontinued, "product is discontinued")
		errsRes = append(errsRes, parserErr)
		return nil, errsRes
	} else {
		status = enums.PRODUCT_STATUS_ACTIVE
	}

	desc := row[idxDesc]

	errProductName := utils.ValidateNotBlank(name, "article")
	if errProductName != nil {
		parserErr := errs.NewParserError(errs.BlankProductName, errProductName.Error())
		errsRes = append(errsRes, parserErr)
	}

	prices, errMonies := DeltaPlusProcessPrices(tpl, row)
	if len(errMonies) > 0 {
		errsRes = append(errsRes, errMonies...)
	}

	if len(errsRes) > 0 {
		return nil, errsRes
	}

	return &models.Product{
		Name:                name,
		Description:         desc,
		Brand:               tpl.Brand,
		Status:              status,
		UnitPriceWithoutVat: prices[0],
		UnitPriceWithVat:    prices[1],
	}, nil
}

func DeltaPlusRowToSpecs(tpl *Template, row []string) *models.ProductSpecs {
	idxColours := tpl.Columns["COLOURS"].Index
	idxSizes := tpl.Columns["SIZES"].Index
	idxSegmentation := tpl.Columns["SEGMENTATION"].Index
	colours := utils.SanitizeColours(row[idxColours])
	sizes := utils.SanitizeSize(row[idxSizes])
	segmentation := row[idxSegmentation]
	return &models.ProductSpecs{
		Colours:      colours,
		Sizes:        sizes,
		Segmentation: segmentation,
	}
}

func DeltaPlusProcessRows(tpl *Template, rows *excelize.Rows) []*models.Product {
	var products []*models.Product = make([]*models.Product, 0, tpl.AssumedRowsCount)

	rowIdx := 0
	for i := 0; i < tpl.SkipInitialRows+1; i++ {
		rows.Next()
		rowIdx++
	}

	totalErrors := 0

	var previousRow []string
	category := ""
	subcategory := ""

LoopProductProces:
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

		if len(row) == 1 {
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
		product, errsRes := tpl.RowToProduct(tpl, row)

		if product != nil {
			specs := tpl.RowToSpecs(tpl, row)
			product.ProductSpecs = specs
		}

		if len(errsRes) > 0 {
			proceedToError := true
			// duplicate := false

			for i, err := range errsRes {
				pc := errs.ParserErrorCodeCode(err)
				if pc == errs.ProductDiscontinued {
					continue LoopProductProces
				}
				if pc == errs.BlankProductName {
					// duplicate = true
					errsRes[i] = nil
					break
				}
			}

			// if duplicate {
			// 	name := row[tpl.Columns["ARTICLE"].Index]
			// 	if name == "" {
			// 		sizes := row[tpl.Columns["SIZES"].Index]
			// 		if sizes != "" {
			// 			prevProduct := products[len(products)-1]
			// 			product = prevProduct.Duplicate()
			// 			product.ProductSpecs.Sizes = sizes
			//
			// 			prices, _ := DeltaPlusProcessPrices(tpl, row)
			// 			if len(prices) > 0 {
			// 				product.UnitPriceWithoutVat = prices[0]
			// 				product.UnitPriceWithVat = prices[1]
			// 			}
			// 			proceedToError = false
			// 		}
			// 	}
			// }

			if proceedToError {
				if tpl.AppFlags.Strict {
					logs.Log().Panic("error", zap.Errors("errors", errsRes))
					return nil
				}
				logs.Log().Debug(
					"row to product",
					zap.Int("row number", rowIdx),
					zap.Errors("errors", errsRes),
				)
				totalErrors += 1
				previousRow = row
				continue LoopProductProces
			}
		}

		product.ProductCategory = &models.ProductCategory{
			Category:    category,
			Subcategory: subcategory,
		}
		product.PostProcess(rowIdx)

		if (tpl.AppFlags.Limit > 0) && (rowIdx > tpl.AppFlags.Limit) {
			return products
		}

		products = append(products, product)
		previousRow = row
	}

	logs.Log().Info(
		"results",
		zap.Int("processed", len(products)),
		zap.Int("errors", totalErrors),
		zap.Int("total", len(products)+totalErrors),
	)

	return products
}

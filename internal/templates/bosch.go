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

var BoschColumns map[string]*Column = map[string]*Column{
	"Model": {
		Index:    -1,
		Required: true,
	},
	"Category": {
		Index:    -1,
		Required: true,
	},
	"Part Number": {
		Index:    -1,
		Required: true,
	},
	"Power": {
		Index:    -1,
		Required: true,
	},
	"Capacity": {
		Index:    -1,
		Required: true,
	},
	"Scope of supply": {
		Index:    -1,
		Required: true,
	},
	"Retail Price (Local Currency)": {
		Index:    -1,
		Required: true,
	},
}

func BoschProcessPrices(tpl *Template, row []string) ([]*money.Money, []error) {
	idxPriceWithoutVat := tpl.Columns["Retail Price (Local Currency)"].Index
	idxPriceWithVat := tpl.Columns["Retail Price (Local Currency)"].Index
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

func BoschRowToProduct(tpl *Template, row []string) (*models.Product, []error) {
	errsRes := make([]error, 0, 8)

	idxModel := tpl.Columns["Model"].Index

	name := row[idxModel]
	errProductName := utils.ValidateNotBlank(name, "Model")
	if errProductName != nil {
		parserErr := errs.NewParserError(errs.BlankProductName, errProductName.Error())
		errsRes = append(errsRes, parserErr)
	}

	prices, errMonies := BoschProcessPrices(tpl, row)
	if len(errMonies) > 0 {
		errsRes = append(errsRes, errMonies...)
	}

	if len(errsRes) > 0 {
		return nil, errsRes
	}

	status := enums.PRODUCT_STATUS_ACTIVE

	descriptions := []string{
		strings.TrimSpace(row[tpl.Columns["Part Number"].Index]),
		strings.TrimSpace(row[tpl.Columns["Power"].Index]),
		strings.TrimSpace(row[tpl.Columns["Capacity"].Index]),
		strings.TrimSpace(row[tpl.Columns["Scope of supply"].Index]),
	}
	descriptions = utils.RemoveEmptyStrings(descriptions)

	return &models.Product{
		Name:                name,
		Brand:               tpl.Brand,
		Description:         strings.Join(descriptions, " - "),
		Status:              status,
		UnitPriceWithoutVat: prices[0],
		UnitPriceWithVat:    prices[1],
	}, nil
}

func BoschRowToSpecs(tpl *Template, row []string) *models.ProductSpecs {
	idxPartNumber := tpl.Columns["Part Number"].Index
	idxPower := tpl.Columns["Power"].Index
	idxCapacity := tpl.Columns["Capacity"].Index
	idxScopeOfSupply := tpl.Columns["Scope of supply"].Index

	partNumber := row[idxPartNumber]
	power := row[idxPower]
	capacity := row[idxCapacity]
	scopeOfSupply := row[idxScopeOfSupply]

	return &models.ProductSpecs{
		PartNumber:    partNumber,
		Power:         power,
		Capacity:      capacity,
		ScopeOfSupply: scopeOfSupply,
	}
}

func BoschProcessRows(tpl *Template, rows *excelize.Rows) []*models.Product {
	var products []*models.Product = make([]*models.Product, 0, tpl.AssumedRowsCount)

	rowIdx := 0
	for i := 0; i < tpl.SkipInitialRows+1; i++ {
		rows.Next()
		rowIdx++
	}

	totalErrors := 0

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

		row = tpl.AlignRow(row)
		product, errs := tpl.RowToProduct(tpl, row)
		specs := tpl.RowToSpecs(tpl, row)

		if len(errs) > 0 {
			proceedToError := true
			if proceedToError {
				if tpl.AppFlags.Strict {
					logs.Log().Panic("error", zap.Errors("errors", errs))
					return nil
				}
				logs.Log().Debug(
					"row to product",
					zap.Int("row number", rowIdx),
					zap.Errors("errors", errs),
				)
				totalErrors += 1
				continue LoopProductProces
			}
		}

		idxCategory := tpl.Columns["Category"].Index
		category := row[idxCategory]

		product.ProductCategory = &models.ProductCategory{
			Category: category,
		}
		product.ProductSpecs = specs
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

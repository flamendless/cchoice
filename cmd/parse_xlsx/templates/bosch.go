package templates

import (
	"cchoice/cmd/parse_xlsx/models"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
	"fmt"
	"os"
	"slices"
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
	colRP, ok := tpl.Columns["Retail Price (Local Currency)"]
	if !ok {
		panic("missing column")
	}
	priceWithoutVat := row[colRP.Index]
	priceWithVat := row[colRP.Index]

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
	colPartNumber, ok := tpl.Columns["Part Number"]
	if !ok {
		panic("missing column")
	}

	colPower, ok := tpl.Columns["Power"]
	if !ok {
		panic("missing column")
	}

	colCapacity, ok := tpl.Columns["Capacity"]
	if !ok {
		panic("missing column")
	}

	colScopeOfSupply, ok := tpl.Columns["Scope of supply"]
	if !ok {
		panic("missing column")
	}

	colModel, ok := tpl.Columns["Model"]
	if !ok {
		panic("missing column")
	}

	errsRes := make([]error, 0, 8)
	name := row[colModel.Index]
	errProductName := utils.ValidateNotBlank(name, "Model")
	if errProductName != nil {
		parserErr := errs.NewParserError(errs.BlankProductName, errProductName.Error())
		errsRes = append(errsRes, parserErr)
	}

	prices, errMonies := BoschProcessPrices(tpl, row)
	if len(prices) < 2 {
		panic("invalid prices")
	}
	if len(errMonies) > 0 {
		errsRes = append(errsRes, errMonies...)
	}

	if len(errsRes) > 0 {
		return nil, errsRes
	}

	status := enums.PRODUCT_STATUS_ACTIVE

	descriptions := []string{
		strings.TrimSpace(row[colPartNumber.Index]),
		strings.TrimSpace(row[colPower.Index]),
		strings.TrimSpace(row[colCapacity.Index]),
		strings.TrimSpace(row[colScopeOfSupply.Index]),
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
	colPartNumber, ok := tpl.Columns["Part Number"]
	if !ok {
		panic("missing column")
	}

	colPower, ok := tpl.Columns["Power"]
	if !ok {
		panic("missing column")
	}

	colCapacity, ok := tpl.Columns["Capacity"]
	if !ok {
		panic("missing column")
	}

	colScopeOfSupply, ok := tpl.Columns["Scope of supply"]
	if !ok {
		panic("missing column")
	}

	partNumber := row[colPartNumber.Index]
	power := row[colPower.Index]
	capacity := row[colCapacity.Index]
	scopeOfSupply := row[colScopeOfSupply.Index]

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
		if product == nil {
			continue
		}
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

		colCategory, ok := tpl.Columns["Category"]
		if !ok {
			panic("missing column")
		}
		category := utils.SanitizeCategory(row[colCategory.Index])
		subcategory := category

		keywords := strings.Split(category, " ")
		if len(keywords) > 1 {
			idx := len(keywords) - 1
			category = keywords[idx]
			keywords = slices.Delete(keywords, idx, idx+1)
			subcategory = strings.Join(keywords, " ")
		}

		product.ProductCategory = &models.ProductCategory{
			Category:    utils.SanitizeCategory(category),
			Subcategory: utils.SanitizeCategory(subcategory),
		}
		if product.ProductCategory.Category == "" {
			panic(fmt.Sprintf("product '%s' has no category value", product.Name))
		}
		if product.ProductCategory.Subcategory == "" {
			panic(fmt.Sprintf("product '%s' has no subcategory value", product.Name))
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

func BoschProcessProductImage(tpl *Template, product *models.Product) (*models.ProductImage, error) {
	basepath := strings.TrimPrefix(tpl.AppFlags.ImagesBasePath, "./cmd/web/")
	basepath = strings.TrimSuffix(basepath, "/")
	filename := strings.ReplaceAll(product.Name, " ", "_")
	path := fmt.Sprintf("%s/%s.png", basepath, filename)

	imagePath := "./cmd/web/" + path
	_, err := os.Stat(imagePath)
	if err != nil {
		logs.Log().Debug("Image path does not exists: " + imagePath)
		return nil, err
	}

	res := &models.ProductImage{
		Product: product,
		Path:    path,
	}
	return res, nil
}

func BoschGetPromotedCategories() []string {
	promoted := []string{
		"chargers",
		"cleaner",
		"cutter",
		"drills",
		"drivers",
		"grinders",
		"hammers",
		"mixer",
		"sander",
		"saws",
		"solution",
		"system",
		"wrench",
	}
	return promoted
}

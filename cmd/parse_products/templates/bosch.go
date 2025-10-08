package templates

import (
	"cchoice/cmd/parse_products/models"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
	"fmt"
	"os"
	"slices"
	"strconv"
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
	"Weight": {
		Index:    -1,
		Required: true,
	},
}

func BoschProcessPrices(tpl *Template, row []string) ([]*money.Money, []error) {
	colRP, ok := tpl.Columns["Retail Price (Local Currency)"]
	if !ok {
		panic(errs.ErrCmdMissingColumn)
	}
	priceWithoutVat := row[colRP.Index]
	priceWithVat := row[colRP.Index]

	prices := make([]*money.Money, 0, 8)
	errors := make([]error, 0, 8)

	unitPriceWithoutVat, err := utils.SanitizePrice(priceWithoutVat)
	if err != nil {
		errors = append(errors, err...)
	}

	unitPriceWithVat, err := utils.SanitizePrice(priceWithVat)
	if err != nil {
		errors = append(errors, err...)
	}

	if len(errors) == 0 {
		prices = append(prices, unitPriceWithoutVat)
		prices = append(prices, unitPriceWithVat)
	}

	return prices, errors
}

func BoschRowToProduct(tpl *Template, row []string) (*models.Product, []error) {
	colPartNumber, ok := tpl.Columns["Part Number"]
	if !ok {
		panic(errs.ErrCmdMissingColumn)
	}

	colPower, ok := tpl.Columns["Power"]
	if !ok {
		panic(errs.ErrCmdMissingColumn)
	}

	colCapacity, ok := tpl.Columns["Capacity"]
	if !ok {
		panic(errs.ErrCmdMissingColumn)
	}

	colScopeOfSupply, ok := tpl.Columns["Scope of supply"]
	if !ok {
		panic(errs.ErrCmdMissingColumn)
	}

	colModel, ok := tpl.Columns["Model"]
	if !ok {
		panic(errs.ErrCmdMissingColumn)
	}

	errsRes := make([]error, 0, 8)
	name := row[colModel.Index]
	errProductName := utils.ValidateNotBlank(name, "Model")
	if errProductName != nil {
		parserErr := errs.NewParserError(errs.BlankProductName, "%s", errProductName.Error())
		errsRes = append(errsRes, parserErr)
	}

	prices, errMonies := BoschProcessPrices(tpl, row)
	if len(prices) < 2 {
		panic(errs.ErrCmdInvalidPrice)
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
		panic(errs.ErrCmdMissingColumn)
	}

	colPower, ok := tpl.Columns["Power"]
	if !ok {
		panic(errs.ErrCmdMissingColumn)
	}

	colCapacity, ok := tpl.Columns["Capacity"]
	if !ok {
		panic(errs.ErrCmdMissingColumn)
	}

	colScopeOfSupply, ok := tpl.Columns["Scope of supply"]
	if !ok {
		panic(errs.ErrCmdMissingColumn)
	}

	colWeight, ok := tpl.Columns["Weight"]
	if !ok {
		panic(errs.ErrCmdMissingColumn)
	}

	partNumber := row[colPartNumber.Index]
	power := row[colPower.Index]
	capacity := row[colCapacity.Index]
	scopeOfSupply := row[colScopeOfSupply.Index]

	weightStr := row[colWeight.Index]
	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		panic(err)
	}

	return &models.ProductSpecs{
		PartNumber:    partNumber,
		Power:         power,
		Capacity:      capacity,
		ScopeOfSupply: scopeOfSupply,
		Weight:        weight,
		WeightUnit:    enums.WEIGHT_UNIT_KG,
	}
}

func BoschProcessRows(tpl *Template, rows *excelize.Rows) []*models.Product {
	products := make([]*models.Product, 0, tpl.AssumedRowsCount)

	rowIdx := 0
	for range tpl.SkipInitialRows + 1 {
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
		product, errsVars := tpl.RowToProduct(tpl, row)
		if product == nil {
			continue
		}
		specs := tpl.RowToSpecs(tpl, row)

		if len(errsVars) > 0 {
			proceedToError := true
			if proceedToError {
				if tpl.AppFlags.Strict {
					logs.Log().Panic("error", zap.Errors("errors", errsVars))
					return nil
				}
				logs.Log().Debug(
					"row to product",
					zap.Int("row number", rowIdx),
					zap.Errors("errors", errsVars),
				)
				totalErrors += 1
				continue LoopProductProces
			}
		}

		colCategory, ok := tpl.Columns["Category"]
		if !ok {
			panic(errs.ErrCmdMissingColumn)
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

		category = utils.SanitizeCategory(category)
		switch category {
		case "GRIDNERS":
			category = "GRINDERS"
		case "PPRUNERS":
			category = "PRUNERS"
		case "CUTERS":
			category = "CUTTERS"
		}

		product.ProductCategory = &models.ProductCategory{
			Category:    category,
			Subcategory: utils.SanitizeCategory(subcategory),
		}
		if product.ProductCategory.Category == "" {
			panic(fmt.Errorf("%w: product '%s'", errs.ErrCmdNoCategory, product.Name))
		}
		if product.ProductCategory.Subcategory == "" {
			panic(fmt.Errorf("%w: product '%s'", errs.ErrCmdNoSubcategory, product.Name))
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
		logs.Log().Debug(
			"Image path does not exists",
			zap.Int64("product id", product.ID),
			zap.String("product name", product.Name),
			zap.String("image path", imagePath),
		)
		return nil, err
	}

	var thumbnail string
	switch tpl.AppFlags.ImagesFormat {
	case "png":
		thumbnail = path
	case "webp":
		path := fmt.Sprintf("%s/640x640/%s.webp", basepath, filename)
		path = strings.Replace(path, "original", "webp", 1)
		thumbnailPath := "./cmd/web/" + path
		_, err := os.Stat(thumbnailPath)
		if err != nil {
			logs.Log().Debug("Image thumbnail path does not exists", zap.String("path", thumbnailPath))
		}
		thumbnail = path
	}

	res := &models.ProductImage{
		Product:   product,
		Path:      path,
		Thumbnail: thumbnail,
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

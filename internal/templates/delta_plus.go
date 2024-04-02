package templates

import (
	"cchoice/internal/models"
	"cchoice/utils"
	"errors"
	"fmt"

	"github.com/xuri/excelize/v2"
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
	// idxPriceWithoutVat := tpl.Columns["END USER UNIT PRICE PHP WITHOUT VAT*"].Index
	// idxPriceWithVat := tpl.Columns["END USER UNIT PRICE PHP WITH VAT* (SRP)"].Index

	var errs error

	name := row[idxArticle]
	errProductName := utils.ValidateNotBlank(name, "article")
	if errProductName != nil {
		errs = errors.Join(errs, errProductName)
	}

	colours := row[idxColours]
	sizes := row[idxSizes]
	if sizes == "-" {
		sizes = ""
	}
	segmentation := row[idxSegmentation]
	desc := row[idxDesc]

	if errs != nil {
		return nil, errs
	}

	return &models.Product{
		Name:         name,
		Description:  desc,
		Colours:      colours,
		Sizes:        sizes,
		Segmentation: segmentation,
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

	rows.Next()

	rowIdx := 0

	var previousRow []string
	category := ""
	subcategory := ""

	for rows.Next() {
		rowIdx++
		row, err := rows.Columns()
		if err != nil {
			fmt.Println(err)
			return products
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

		fmt.Println(rowIdx, category, subcategory)

		if len(row) == 0 {
			break
		}

		row = tpl.AlignRow(row)
		product, errs := tpl.RowToProduct(tpl, row)
		if errs != nil {
			if tpl.Flags.Strict {
				fmt.Println(errs)
				panic("error immediately")
			}
			fmt.Printf("row %d: %s\n", rowIdx, errs)
			previousRow = row
			continue
		}

		product.Category = category
		product.Subcategory = subcategory

		if (tpl.Flags.Limit != 0) && (rowIdx > tpl.Flags.Limit) {
			return products
		}

		products = append(products, product)
		previousRow = row
	}

	return products
}

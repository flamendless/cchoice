package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"

	"cchoice/internal/models"
	"cchoice/internal/templates"
)

var (
	template string
	filepath string
	sheet    string
	strict   bool
)

func init() {
	parseXLSXCmd.Flags().StringVarP(&template, "template", "t", "", "Template to use")
	parseXLSXCmd.Flags().StringVarP(&filepath, "filepath", "p", "", "Filepath to the XLSX file")
	parseXLSXCmd.Flags().StringVarP(&sheet, "sheet", "s", "", "Sheet name to use")
	parseXLSXCmd.Flags().BoolVarP(&strict, "strict", "x", false, "Panic upon first product error")

	parseXLSXCmd.MarkFlagRequired("template")
	parseXLSXCmd.MarkFlagRequired("filepath")

	rootCmd.AddCommand(parseXLSXCmd)
}

func ProcessColumns(tpl *templates.Template, file *excelize.File, sheet string) bool {
	rows, err := file.Rows(sheet)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	rows.Next()
	columns, err := rows.Columns()
	if err != nil {
		fmt.Println(err)
		return false
	}

	for i, cell := range columns {
		col, _ := tpl.Columns[cell]
		col.Index = i
	}

	return tpl.ValidateColumns()
}

func RowsToProducts(tpl *templates.Template, file *excelize.File, sheet string) []*models.Product {
	var products []*models.Product = make([]*models.Product, 0, tpl.AssumedRowsCount)

	rows, err := file.Rows(sheet)
	if err != nil {
		fmt.Println(err)
		return products
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

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
			if strict {
				fmt.Println(errs)
				panic("error immediately")
			}
			fmt.Printf("row %d: %s\n", rowIdx, errs)
			continue
		}
		products = append(products, product)
	}

	return products
}

var parseXLSXCmd = &cobra.Command{
	Use:   "parse_xlsx",
	Short: "Parse XLSX file",
	Run: func(cmd *cobra.Command, args []string) {
		templateKind := templates.ParseTemplateEnum(template)
		tpl := templates.CreateTemplate(templateKind)

		file, err := excelize.OpenFile(filepath)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func() {
			err = file.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
		}()

		if sheet == "" {
			sheet = file.GetSheetName(0)
		}

		success := ProcessColumns(tpl, file, sheet)
		if !success {
			return
		}

		fmt.Println()

		products := RowsToProducts(tpl, file, sheet)
		fmt.Println()
		fmt.Println("Products length:", len(products))
		for _, product := range products {
			product.Print()
		}
	},
}

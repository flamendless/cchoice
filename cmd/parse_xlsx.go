package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"

	"cchoice/internal/templates"
)

var (
	template string
	filepath string
	sheet    string
	strict   bool
	limit    int
)

func init() {
	parseXLSXCmd.Flags().StringVarP(&template, "template", "t", "", "Template to use")
	parseXLSXCmd.Flags().StringVarP(&filepath, "filepath", "p", "", "Filepath to the XLSX file")
	parseXLSXCmd.Flags().StringVarP(&sheet, "sheet", "s", "", "Sheet name to use")
	parseXLSXCmd.Flags().BoolVarP(&strict, "strict", "x", false, "Panic upon first product error")
	parseXLSXCmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit number of rows to process")

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
	for i := 0; i < tpl.SkipInitialRows; i++ {
		rows.Next()
	}

	columns, err := rows.Columns()
	if err != nil {
		fmt.Println(err)
		return false
	}

	for i, cell := range columns {
		cell = strings.Replace(cell, "\n", " ", -1)
		col, exists := tpl.Columns[cell]
		if !exists {
			fmt.Printf("Column '%s' does not exist in template columns\n", cell)
			continue
		}
		col.Index = i
	}

	return tpl.ValidateColumns()
}

var parseXLSXCmd = &cobra.Command{
	Use:   "parse_xlsx",
	Short: "Parse XLSX file",
	Run: func(cmd *cobra.Command, args []string) {
		templateKind := templates.ParseTemplateEnum(template)
		tpl := templates.CreateTemplate(templateKind)
		tpl.Flags.Limit = limit
		tpl.Flags.Strict = strict

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
		tpl.Sheet = sheet

		success := ProcessColumns(tpl, file, sheet)
		if !success {
			return
		}

		fmt.Println()

		rows, err := file.Rows(tpl.Sheet)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func() {
			err := rows.Close()
			if err != nil {
				fmt.Println(err)
			}
		}()

		if tpl.ProcessRows == nil {
			panic(fmt.Sprintf(
				"Must provide template '%s' -> ProcessRows\n",
				templateKind.String(),
			))
		}

		products := tpl.ProcessRows(tpl, rows)
		fmt.Println()
		fmt.Println("Products length:", len(products))
		for _, product := range products {
			product.Print()
		}
	},
}

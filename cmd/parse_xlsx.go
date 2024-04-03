package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"

	"cchoice/internal"
	"cchoice/internal/templates"
)

var ctx internal.AppContext

func init() {
	f := parseXLSXCmd.Flags
	f().StringVarP(&ctx.Template, "template", "t", "", "Template to use")
	f().StringVarP(&ctx.Filepath, "filepath", "p", "", "Filepath to the XLSX file")
	f().StringVarP(&ctx.Sheet, "sheet", "s", "", "Sheet name to use")
	f().BoolVarP(&ctx.Strict, "strict", "x", false, "Panic upon first product error")
	f().IntVarP(&ctx.Limit, "limit", "l", 0, "Limit number of rows to process")

	parseXLSXCmd.MarkFlagRequired("template")
	parseXLSXCmd.MarkFlagRequired("filepath")

	rootCmd.AddCommand(parseXLSXCmd)
}

func ProcessColumns(tpl *templates.Template, file *excelize.File) bool {
	rows, err := file.Rows(tpl.AppContext.Sheet)
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
		templateKind := templates.ParseTemplateEnum(ctx.Template)
		tpl := templates.CreateTemplate(templateKind)

		file, err := excelize.OpenFile(ctx.Filepath)
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

		if ctx.Sheet == "" {
			ctx.Sheet = file.GetSheetName(0)
		}

		tpl.AppContext = &ctx

		success := ProcessColumns(tpl, file)
		if !success {
			return
		}

		fmt.Println()

		rows, err := file.Rows(ctx.Sheet)
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

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	"cchoice/internal"
	"cchoice/internal/logs"
	"cchoice/internal/templates"
)

var ctx internal.AppContext

func init() {
	f := parseXLSXCmd.Flags
	f().StringVarP(&ctx.Template, "template", "t", "", "Template to use")
	f().StringVarP(&ctx.Filepath, "filepath", "p", "", "Filepath to the XLSX file")
	f().StringVarP(&ctx.Sheet, "sheet", "s", "", "Sheet name to use")
	f().BoolVarP(&ctx.Strict, "strict", "x", false, "Panic upon first product error")
	f().BoolVarP(&ctx.PrintProcessedProducts, "print_processed_products", "v", false, "Print processed products")
	f().IntVarP(&ctx.Limit, "limit", "l", 0, "Limit number of rows to process")

	logs.InitLog()

	parseXLSXCmd.MarkFlagRequired("template")
	parseXLSXCmd.MarkFlagRequired("filepath")

	rootCmd.AddCommand(parseXLSXCmd)
}

func ProcessColumns(tpl *templates.Template, file *excelize.File) bool {
	rows, err := file.Rows(tpl.AppContext.Sheet)
	if err != nil {
		logs.Log().Error(err.Error())
		return false
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			logs.Log().Error(err.Error())
		}
	}()

	rows.Next()
	for i := 0; i < tpl.SkipInitialRows; i++ {
		rows.Next()
	}

	columns, err := rows.Columns()
	if err != nil {
		logs.Log().Error(err.Error())
		return false
	}

	for i, cell := range columns {
		cell = strings.Replace(cell, "\n", " ", -1)
		col, exists := tpl.Columns[cell]
		if !exists {
			logs.Log().Info(fmt.Sprintf(
				"Column '%s' does not exist in template columns\n",
				cell,
			))
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
		logs.Log().Info(
			"Parsing XLSX",
			zap.String("template", ctx.Template),
			zap.String("filepath", ctx.Filepath),
			zap.String("sheet", ctx.Sheet),
			zap.Bool("strict", ctx.Strict),
			zap.Int("limit", ctx.Limit),
		)

		templateKind := templates.ParseTemplateEnum(ctx.Template)
		tpl := templates.CreateTemplate(templateKind)

		file, err := excelize.OpenFile(ctx.Filepath)
		if err != nil {
			logs.Log().Error(err.Error())
			return
		}
		defer func() {
			err = file.Close()
			if err != nil {
				logs.Log().Error(err.Error())
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

		rows, err := file.Rows(ctx.Sheet)
		if err != nil {
			logs.Log().Error(err.Error())
			return
		}
		defer func() {
			err := rows.Close()
			if err != nil {
				logs.Log().Error(err.Error())
			}
		}()

		if tpl.ProcessRows == nil {
			logs.Log().Panic(fmt.Sprintf(
				"Must provide template '%s' -> ProcessRows\n",
				templateKind.String(),
			))
			return
		}

		products := tpl.ProcessRows(tpl, rows)

		if tpl.AppContext.PrintProcessedProducts {
			for _, product := range products {
				product.Print()
			}
		}
	},
}

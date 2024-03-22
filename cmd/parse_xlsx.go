package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"

	"cchoice/internal/templates"
)

var (
	template string
	filepath string
	sheet    string
)

func init() {
	parseXLSXCmd.Flags().StringVarP(&template, "template", "t", "", "Template to use")
	parseXLSXCmd.Flags().StringVarP(&filepath, "filepath", "p", "", "Filepath to the XLSX file")
	parseXLSXCmd.Flags().StringVarP(&sheet, "sheet", "s", "", "Sheet name to use")

	parseXLSXCmd.MarkFlagRequired("template")
	parseXLSXCmd.MarkFlagRequired("filepath")

	rootCmd.AddCommand(parseXLSXCmd)
}

func ValidateColumns(tpl *templates.Template, file *excelize.File, sheet string) bool {
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

	for key, col := range tpl.Columns {
		if col.Required && col.Index == -1 {
			fmt.Printf("Failed to find column index for '%s'\n", key)
			return false
		}
	}

	return true
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

		success := ValidateColumns(tpl, file, sheet)
		if !success {
			return
		}

		tpl.Print()
	},
}

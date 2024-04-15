package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	"cchoice/internal"
	"cchoice/internal/cchoice_db"
	"cchoice/internal/logs"
	"cchoice/internal/models"
	"cchoice/internal/templates"
)

var ctx internal.AppFlags

func init() {
	f := parseXLSXCmd.Flags
	f().StringVarP(&ctx.Template, "template", "t", "", "Template to use")
	f().StringVarP(&ctx.Filepath, "filepath", "p", "", "Filepath to the XLSX file")
	f().StringVarP(&ctx.Sheet, "sheet", "s", "", "Sheet name to use")
	f().StringVarP(&ctx.DBPath, "db_path", "", ":memory:", "Path to database")
	f().BoolVarP(&ctx.Strict, "strict", "x", false, "Panic upon first product error")
	f().BoolVarP(&ctx.PrintProcessedProducts, "print_processed_products", "v", false, "Print processed products")
	f().BoolVarP(&ctx.UseDB, "use_db", "", false, "Use DB to save processed data")
	f().BoolVarP(&ctx.VerifyPrices, "verify_prices", "", true, "Verify prices processed and saved to DB")
	f().IntVarP(&ctx.Limit, "limit", "l", 0, "Limit number of rows to process")

	logs.InitLog()

	parseXLSXCmd.MarkFlagRequired("template")
	parseXLSXCmd.MarkFlagRequired("filepath")

	rootCmd.AddCommand(parseXLSXCmd)
}

func ProcessColumns(tpl *templates.Template, file *excelize.File) bool {
	rows, err := file.Rows(tpl.AppFlags.Sheet)
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

		tpl.AppFlags = &ctx
		tpl.AppContext = &internal.AppContext{}

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

		if tpl.AppFlags.PrintProcessedProducts {
			for _, product := range products {
				product.Print()
			}
		}

		if !tpl.AppFlags.UseDB {
			return
		}

		sqlDB, err := cchoicedb.InitDB(tpl.AppFlags.DBPath)
		if err != nil {
			logs.Log().Error(
				"DB initialization",
				zap.Error(err),
			)
			return
		}
		tpl.AppContext.DB = sqlDB
		tpl.AppContext.Queries = cchoicedb.GetQueries(sqlDB)

		logs.Log().Debug("Inserting/updating products to DB...")

		insertedIds := make([]int64, 0, len(products))
		updatedIds := make([]int64, 0, len(products))

		for _, product := range products {
			existingProductId := product.GetDBID(tpl.AppContext)
			if existingProductId != 0 {
				productID, err := product.UpdateToDB(tpl.AppContext)
				if err != nil {
					logs.Log().Info(
						"product update to DB",
						zap.Error(err),
					)
					continue
				}
				updatedIds = append(updatedIds, productID)

			} else {
				productID, err := product.InsertToDB(tpl.AppContext)
				if err != nil {
					logs.Log().Info(
						"product insert to DB",
						zap.Error(err),
					)
					continue
				}
				insertedIds = append(insertedIds, productID)
			}
		}

		if tpl.AppFlags.PrintProcessedProducts {
			logs.Log().Debug(
				"Products inserted/updated to DB",
				zap.Int64s("inserted ids", insertedIds),
				zap.Int64s("updated ids", updatedIds),
			)
		}
		logs.Log().Debug(
			"Successfully inserted/updated products to DB",
			zap.Int("inserted ids count", len(insertedIds)),
			zap.Int("updated ids count", len(updatedIds)),
		)

		if tpl.AppFlags.VerifyPrices {
			logs.Log().Debug("Verifying prices...")
			for i := 0; i < len(products); i++ {
				product := products[i]
				row, _ := tpl.AppContext.Queries.GetProductByID(
					context.Background(),
					int64(i+1),
				)
				dbp := models.DBRowToProduct(&row)

				cmp, _ := product.UnitPriceWithoutVat.Compare(dbp.UnitPriceWithoutVat)
				if cmp != 0 {
					logs.Log().Warn(
						"checking prices",
						zap.Int64("id", product.ID),
						zap.Int64("product", product.UnitPriceWithoutVat.Amount()),
						zap.Int64("db", dbp.UnitPriceWithoutVat.Amount()),
					)
				}
			}
			logs.Log().Debug("Successfully verified prices")
		}
	},
}

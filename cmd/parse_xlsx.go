package cmd

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	cchoicedb "cchoice/internal/cchoice_db"
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	"cchoice/internal/models"
	"cchoice/internal/templates"
)

var ctxParseXLSX ctx.ParseXLSXFlags

func init() {
	f := parseXLSXCmd.Flags
	f().StringVarP(&ctxParseXLSX.Template, "template", "t", "", "Template to use")
	f().StringVarP(&ctxParseXLSX.Filepath, "filepath", "p", "", "Filepath to the XLSX file")
	f().StringVarP(&ctxParseXLSX.Sheet, "sheet", "s", "", "Sheet name to use")
	f().StringVarP(&ctxParseXLSX.DBPath, "db_path", "", ":memory:", "Path to database")
	f().BoolVarP(&ctxParseXLSX.Strict, "strict", "x", false, "Panic upon first product error")
	f().BoolVarP(&ctxParseXLSX.PrintProcessedProducts, "print_processed_products", "v", false, "Print processed products")
	f().BoolVarP(&ctxParseXLSX.UseDB, "use_db", "", false, "Use DB to save processed data")
	f().BoolVarP(&ctxParseXLSX.VerifyPrices, "verify_prices", "", true, "Verify prices processed and saved to DB")
	f().IntVarP(&ctxParseXLSX.Limit, "limit", "l", 0, "Limit number of rows to process")

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
			zap.String("template", ctxParseXLSX.Template),
			zap.String("filepath", ctxParseXLSX.Filepath),
			zap.String("sheet", ctxParseXLSX.Sheet),
			zap.Bool("strict", ctxParseXLSX.Strict),
			zap.Int("limit", ctxParseXLSX.Limit),
		)

		templateKind := templates.ParseTemplateEnum(ctxParseXLSX.Template)
		tpl := templates.CreateTemplate(templateKind)

		file, err := excelize.OpenFile(ctxParseXLSX.Filepath)
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

		if ctxParseXLSX.Sheet == "" {
			ctxParseXLSX.Sheet = file.GetSheetName(0)
		}

		tpl.AppFlags = &ctxParseXLSX
		tpl.AppContext = &ctx.App{}
		tpl.AppContext.Metrics = &ctx.Metrics{}

		startProcessColumns := time.Now()
		success := ProcessColumns(tpl, file)
		if !success {
			return
		}
		tpl.AppContext.Metrics.Add("process column time", time.Since(startProcessColumns))

		rows, err := file.Rows(ctxParseXLSX.Sheet)
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

		startProcessRows := time.Now()
		products := tpl.ProcessRows(tpl, rows)
		tpl.AppContext.Metrics.Add("process rows time", time.Since(startProcessRows))

		if tpl.AppFlags.PrintProcessedProducts {
			for _, product := range products {
				product.Print()
			}
		}

		if !tpl.AppFlags.UseDB {
			tpl.AppContext.Metrics.LogTime(logs.Log())
			return
		}

		sqlDBRead, err := cchoicedb.InitDB(tpl.AppFlags.DBPath, "ro")
		if err != nil {
			logs.Log().Error(
				"DB (read-only) initialization",
				zap.Error(err),
			)
			return
		}
		tpl.AppContext.DBRead = sqlDBRead
		tpl.AppContext.QueriesRead = cchoicedb.GetQueries(sqlDBRead)

		sqlDB, err := cchoicedb.InitDB(tpl.AppFlags.DBPath, "rw")
		defer sqlDB.Close()
		if err != nil {
			logs.Log().Error(
				"DB (read-write) initialization",
				zap.Error(err),
			)
			return
		}
		sqlDB.SetMaxOpenConns(1)
		tpl.AppContext.DB = sqlDB
		tpl.AppContext.Queries = cchoicedb.GetQueries(sqlDB)

		logs.Log().Debug("Inserting/updating products to DB...")

		insertedIds := make([]int64, 0, len(products))
		updatedIds := make([]int64, 0, len(products))

		var insertMetrics int64
		var updateMetrics int64

		nProcessors := runtime.NumCPU()
		batchsize := nProcessors*2 - 1
		var wg sync.WaitGroup

		startWG := time.Now()
		for start, end := 0, 0; start <= len(products)-1; start = end {
			end = start + batchsize
			if end > len(products) {
				end = len(products)
			}

			batch := products[start:end]
			wg.Add(1)

			go func() {
				defer wg.Done()
				for _, product := range batch {
					existingProductId := product.GetDBID(tpl.AppContext)
					if existingProductId != 0 {
						now := time.Now()
						_, err := product.UpdateToDB(tpl.AppContext)
						if err != nil {
							logs.Log().Info(
								"product update to DB",
								zap.Int64("id", product.ID),
								zap.Error(err),
							)
						} else {
							updatedIds = append(updatedIds, existingProductId)
							updateMetrics += int64(time.Since(now))
						}

					} else {
						now := time.Now()
						productID, err := product.InsertToDB(tpl.AppContext)
						if err != nil {
							logs.Log().Info(
								"product insert to DB",
								zap.Int64("id", product.ID),
								zap.Error(err),
							)
						} else {
							insertedIds = append(insertedIds, productID)
							insertMetrics += int64(time.Since(now))
						}
					}
				}
			}()
		}
		wg.Wait()
		tpl.AppContext.Metrics.Add("Get product IDS (parallel)", time.Since(startWG))

		logs.Log().Info(
			"parallel processing",
			zap.Int("products", len(products)),
			zap.Int("inserted", len(insertedIds)),
			zap.Int("updated", len(updatedIds)),
		)

		if len(insertedIds) > 0 {
			tpl.AppContext.Metrics.Add(
				"insert products to DB time",
				time.Duration(insertMetrics/int64(len(insertedIds))),
			)
		}
		if len(updatedIds) > 0 {
			tpl.AppContext.Metrics.Add(
				"update products to DB time",
				time.Duration(updateMetrics/int64(len(updatedIds))),
			)
		}

		if tpl.AppFlags.VerifyPrices {
			hasError := false
			logs.Log().Debug("Verifying prices...")
			for i := 0; i < len(products); i++ {
				product := products[i]
				if product.ID == 0 {
					continue
				}

				row, err := tpl.AppContext.Queries.GetProductBySerial(context.Background(), product.Serial)
				if err != nil {
					continue
				}

				dbp := models.DBRowToProduct(&row)
				cmp, _ := product.UnitPriceWithoutVat.Compare(dbp.UnitPriceWithoutVat)
				if cmp != 0 {
					hasError = true
					logs.Log().Warn(
						"checking prices",
						zap.Int64("product ID", product.ID),
						zap.String("product brand", product.Brand),
						zap.String("product serial", product.Serial),
						zap.Int64("product price", product.UnitPriceWithoutVat.Amount()),
						zap.Int64("db ID", dbp.ID),
						zap.String("db brand", dbp.Brand),
						zap.String("db serial", dbp.Serial),
						zap.Int64("db price", dbp.UnitPriceWithoutVat.Amount()),
					)
				}
			}

			if !hasError {
				logs.Log().Debug("Successfully verified prices")
			}
		}

		logs.Log().Debug(
			"Successfully inserted/updated products to DB",
			zap.Int("inserted ids count", len(insertedIds)),
			zap.Int("updated ids count", len(updatedIds)),
		)

		tpl.AppContext.Metrics.LogTime(logs.Log())
	},
}

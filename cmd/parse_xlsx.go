package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	"cchoice/cmd/parse_xlsx/models"
	"cchoice/cmd/parse_xlsx/templates"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/logs"
)

var parseXLSXFlags models.ParseXLSXFlags

func init() {
	f := cmdParseXLSX.Flags
	f().StringVarP(&parseXLSXFlags.Template, "template", "t", "", "Template to use")
	f().StringVarP(&parseXLSXFlags.Filepath, "filepath", "p", "", "Filepath to the XLSX file")
	f().StringVarP(&parseXLSXFlags.Sheet, "sheet", "s", "", "Sheet name to use")
	f().StringVarP(&parseXLSXFlags.DBPath, "db_path", "", ":memory:", "Path to database")
	f().StringVarP(&parseXLSXFlags.ImagesBasePath, "images_basepath", "", "", "Base path to the images")
	f().BoolVarP(&parseXLSXFlags.Strict, "strict", "x", false, "Panic upon first product error")
	f().BoolVarP(&parseXLSXFlags.PrintProcessedProducts, "print_processed_products", "v", false, "Print processed products")
	f().BoolVarP(&parseXLSXFlags.UseDB, "use_db", "", false, "Use DB to save processed data")
	f().BoolVarP(&parseXLSXFlags.VerifyPrices, "verify_prices", "", true, "Verify prices processed and saved to DB")
	f().BoolVarP(&parseXLSXFlags.PanicOnFirstDBError, "panic_on_error", "", false, "Whether to panic immediately on first DB error or not")
	f().IntVarP(&parseXLSXFlags.Limit, "limit", "l", 0, "Limit number of rows to process")

	if err := cmdParseXLSX.MarkFlagRequired("template"); err != nil {
		panic(err)
	}
	if err := cmdParseXLSX.MarkFlagRequired("filepath"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(cmdParseXLSX)
}

func ProcessColumns(tpl *templates.Template, file *excelize.File) bool {
	if tpl.BrandOnly {
		return true
	}

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

var cmdParseXLSX = &cobra.Command{
	Use:   "parse_xlsx",
	Short: "Parse XLSX file",
	Run: func(cmd *cobra.Command, args []string) {
		logs.Log().Info(
			"Parsing XLSX",
			zap.String("template", parseXLSXFlags.Template),
			zap.String("filepath", parseXLSXFlags.Filepath),
			zap.String("sheet", parseXLSXFlags.Sheet),
			zap.Bool("strict", parseXLSXFlags.Strict),
			zap.Int("limit", parseXLSXFlags.Limit),
		)

		templateKind := templates.ParseTemplateEnum(parseXLSXFlags.Template)
		tpl := templates.CreateTemplate(templateKind)

		file, err := excelize.OpenFile(parseXLSXFlags.Filepath)
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

		if parseXLSXFlags.Sheet == "" {
			parseXLSXFlags.Sheet = file.GetSheetName(0)
		}

		tpl.AppFlags = &parseXLSXFlags
		tpl.CtxApp = &models.ParseXLSX{}
		tpl.CtxApp.Metrics = &models.Metrics{}

		startProcessColumns := time.Now()
		success := ProcessColumns(tpl, file)
		if !success {
			return
		}
		tpl.CtxApp.Metrics.Add("process column time", time.Since(startProcessColumns))

		rows, err := file.Rows(parseXLSXFlags.Sheet)
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

		if tpl.ProcessRows == nil && !tpl.BrandOnly {
			logs.Log().Panic(fmt.Sprintf(
				"Must provide template '%s' -> ProcessRows\n",
				templateKind.String(),
			))
			return
		}

		if !tpl.AppFlags.UseDB {
			tpl.CtxApp.Metrics.LogTime(logs.Log())
			return
		}

		tpl.CtxApp.DB = database.New(database.DB_MODE_RW)
		defer func() {
			if err := tpl.CtxApp.DB.Close(); err != nil {
				panic(err)
			}
		}()

		logs.Log().Debug("Getting brand...")
		brand := models.NewBrand(parseXLSXFlags.Template)
		brandID := brand.GetDBID(cmd.Context(), tpl.CtxApp.DB)
		if brandID == 0 {
			brandID, err := brand.InsertToDB(cmd.Context(), tpl.CtxApp.DB)
			if err != nil {
				panic(err)
			}
			now := time.Now().UTC()
			brandImage := models.BrandImage{
				BrandID:   brandID,
				Path:      "static/images/brand_logos/" + parseXLSXFlags.Template + ".png",
				IsMain:    true,
				CreatedAt: now,
				UpdatedAt: now,
				DeletedAt: constants.DT_BEGINNING,
			}
			_, err = brandImage.InsertToDB(cmd.Context(), tpl.CtxApp.DB)
			if err != nil {
				panic(err)
			}
		}
		tpl.Brand = brand
		if tpl.BrandOnly {
			return
		}

		startProcessRows := time.Now()
		products := tpl.ProcessRows(tpl, rows)
		tpl.CtxApp.Metrics.Add("process rows time", time.Since(startProcessRows))

		if tpl.AppFlags.PrintProcessedProducts {
			for _, product := range products {
				product.Print()
			}
		}

		logs.Log().Debug("Inserting/updating products to DB...")

		insertedIds := make([]int64, 0, len(products))
		updatedIds := make([]int64, 0, len(products))

		var insertMetrics int64
		var updateMetrics int64

		nProcessors := runtime.NumCPU()
		batchsize := nProcessors*2 - 1
		var wg sync.WaitGroup

		startWG := time.Now()
		foundImages := 0
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
					existingProductId := product.GetDBID(cmd.Context(), tpl.CtxApp.DB)
					if existingProductId != 0 {
						now := time.Now()
						_, err := product.UpdateToDB(cmd.Context(), tpl.CtxApp.DB)
						if err != nil {
							logs.Log().Info(
								"product update to DB",
								zap.Int64("id", product.ID),
								zap.Error(err),
							)

							if tpl.AppFlags.PanicOnFirstDBError {
								panic(1)
							}
						} else {
							updatedIds = append(updatedIds, existingProductId)
							updateMetrics += int64(time.Since(now))
						}

					} else {
						now := time.Now()
						productID, err := product.InsertToDB(cmd.Context(), tpl.CtxApp.DB)
						if err != nil {
							logs.Log().Info(
								"product insert to DB",
								zap.Int64("id", product.ID),
								zap.Error(err),
							)

							if tpl.AppFlags.PanicOnFirstDBError {
								panic(1)
							}
						} else {
							insertedIds = append(insertedIds, productID)
							insertMetrics += int64(time.Since(now))

							if tpl.AppFlags.ImagesBasePath != "" {
								productImageID, err := ProcessProductImage(cmd.Context(), tpl, product)
								if err != nil && !errors.Is(err, os.ErrNotExist) {
									logs.Log().Info(
										"product image insert to DB",
										zap.Int64("id", productImageID),
										zap.Error(err),
									)
									if tpl.AppFlags.PanicOnFirstDBError {
										panic(1)
									}
								}
								if err == nil {
									foundImages++
								}
							}
						}
					}
				}
			}()
		}
		wg.Wait()
		tpl.CtxApp.Metrics.Add("Get product IDS (parallel)", time.Since(startWG))

		logs.Log().Info(
			"parallel processing",
			zap.Int("products", len(products)),
			zap.Int("inserted", len(insertedIds)),
			zap.Int("updated", len(updatedIds)),
		)

		if len(insertedIds) > 0 {
			tpl.CtxApp.Metrics.Add(
				"insert products to DB time",
				time.Duration(insertMetrics/int64(len(insertedIds))),
			)
		}
		if len(updatedIds) > 0 {
			tpl.CtxApp.Metrics.Add(
				"update products to DB time",
				time.Duration(updateMetrics/int64(len(updatedIds))),
			)
		}
		logs.Log().Info("Images", zap.Int("found images", foundImages), zap.Int("products count", len(products)))

		if tpl.AppFlags.VerifyPrices {
			hasError := false
			logs.Log().Debug("Verifying prices...")
			for i := 0; i < len(products); i++ {
				product := products[i]
				if product.ID == 0 {
					continue
				}

				row, err := tpl.CtxApp.DB.GetQueries().GetProductBySerial(cmd.Context(), product.Serial)
				if err != nil {
					continue
				}

				dbp := models.DBRowToProduct(&row)

				brand := models.NewBrand(row.BrandName)
				_ = brand.GetDBID(cmd.Context(), tpl.CtxApp.DB)
				dbp.Brand = brand

				cmp, _ := product.UnitPriceWithoutVat.Compare(dbp.UnitPriceWithoutVat)
				if cmp != 0 {
					hasError = true
					logs.Log().Warn(
						"checking prices",
						zap.Int64("product ID", product.ID),
						zap.String("product brand", product.Brand.Name),
						zap.String("product serial", product.Serial),
						zap.Int64("product price", product.UnitPriceWithoutVat.Amount()),
						zap.Int64("db ID", dbp.ID),
						zap.String("db brand", dbp.Brand.Name),
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

		params := make([]sql.NullString, 0, len(tpl.GetPromotedCategories()))
		for _, v := range tpl.GetPromotedCategories() {
			params = append(params, sql.NullString{String: v, Valid: true})

		}

		promotedCategoryIDs, err := tpl.CtxApp.DB.GetQueries().SetInitialPromotedProductCategory(cmd.Context(), params)
		if err != nil {
			panic(err)
		}
		logs.Log().Debug(
			"Set initial promoted categories",
			zap.Int("promoted categories count", len(promotedCategoryIDs)),
		)

		tpl.CtxApp.Metrics.LogTime(logs.Log())
	},
}

func ProcessProductImage(ctx context.Context, tpl *templates.Template, product *models.Product) (int64, error) {
	productImage, err := tpl.ProcessProductImage(tpl, product)
	if err != nil {
		return 0, err
	}
	productImageID, err := productImage.InsertToDB(ctx, tpl.CtxApp.DB)
	if err != nil {
		return 0, err
	}
	return productImageID, err
}

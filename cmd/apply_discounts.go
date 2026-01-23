package cmd

import (
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"database/sql"
	"encoding/csv"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	flagInputFile string
	flagDryRun    bool
)

func init() {
	f := cmdApplyDiscount.Flags
	f().StringVarP(&flagInputFile, "input_file", "i", "", "Input filename (.csv)")
	f().BoolVarP(&flagDryRun, "dry-run", "d", true, "Dry run")
	if err := cmdApplyDiscount.MarkFlagRequired("input_file"); err != nil {
		panic(err)
	}
	if err := cmdApplyDiscount.MarkFlagRequired("dry-run"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(cmdApplyDiscount)
}

var cmdApplyDiscount = &cobra.Command{
	Use:   "apply_discount",
	Short: "Apply discount to selected models via CSV file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagInputFile == "" {
			panic(errs.ErrCmdRequired)
		}

		logs.Log().Info(
			"Parameters",
			zap.String("input", flagInputFile),
			zap.Bool("dry run", flagDryRun),
		)

		f, err := os.Open(flagInputFile)
		if err != nil {
			return err
		}
		defer f.Close()

		r := csv.NewReader(f)
		r.TrimLeadingSpace = true

		records, err := r.ReadAll()
		if err != nil {
			return err
		}

		if len(records) <= 1 {
			return errors.New("csv has no data rows")
		}

		header := records[0]
		col := map[string]int{}
		for i, h := range header {
			col[strings.ToLower(h)] = i
		}

		for i, row := range records[1:] {
			line := i + 2
			name := row[col["name"]]
			rawDiscount := row[col["discount"]]

			pd, err := parseDiscount(rawDiscount)
			if err != nil {
				logs.Log().Error(
					"Parsing discount",
					zap.Int("line", line),
					zap.String("name", name),
					zap.String("discount", rawDiscount),
					zap.Error(err),
				)
				continue
			}

			if pd.ShouldSkip {
				logs.Log().Warn(
					"Skipped",
					zap.Int("line", line),
					zap.String("name", name),
					zap.String("discount", rawDiscount),
				)
				continue
			}

			db := database.New(database.DB_MODE_RW)

			product, err := db.GetQueries().GetProductByName(cmd.Context(), name)
			if errors.Is(err, sql.ErrNoRows) {
				logs.Log().Warn("Not found", zap.String("name", name))
				continue
			}
			if err != nil {
				return err
			}

			oldPrice := product.UnitPriceWithVat
			newPrice := oldPrice

			switch pd.Kind {
			case DiscountPercentage:
				if pd.IsSubtract {
					newPrice = oldPrice - (oldPrice*pd.Value)/100
				}
			case DiscountFixed:
				if pd.IsSubtract {
					newPrice = oldPrice - pd.Value
				}
			}

			if newPrice < 0 {
				newPrice = 0
			}

			logs.Log().Info(
				"Update",
				zap.String("name", name),
				zap.Any("old price", oldPrice),
				zap.String("discount", pd.Raw),
				zap.Any("new price", newPrice),
			)

			if !flagDryRun {
				now := time.Now().UTC()
				_, err = db.GetQueries().CreateProductSale(cmd.Context(), queries.CreateProductSaleParams{
					ProductID:                   product.ID,
					SalePriceWithVat:            newPrice,
					SalePriceWithoutVat:         newPrice,
					SalePriceWithVatCurrency:    product.UnitPriceWithVatCurrency,
					SalePriceWithoutVatCurrency: product.UnitPriceWithoutVatCurrency,
					DiscountType:                string(pd.Kind),
					DiscountValue:               pd.Value,
					StartsAt:                    now,
					EndsAt:                      now.AddDate(1, 0, 0),
					IsActive:                    true,
					CreatedAt:                   now,
					UpdatedAt:                   now,
					DeletedAt:                   constants.DtBeginning,
				})
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}

type DiscountRow struct {
	Name     string
	Discount string
}

type DiscountKind string

const (
	DiscountPercentage DiscountKind = "percentage"
	DiscountFixed      DiscountKind = "fixed"
)

type ParsedDiscount struct {
	Kind       DiscountKind
	Value      int64 // percent OR cents
	IsSubtract bool
	ShouldSkip bool
	Raw        string
}

func parseDiscount(raw string) (ParsedDiscount, error) {
	d := ParsedDiscount{Raw: raw}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return d, errors.New("empty discount")
	}

	if strings.HasPrefix(raw, "+") {
		d.ShouldSkip = true
		return d, nil
	}

	if strings.HasPrefix(raw, "-") {
		d.IsSubtract = true
		raw = strings.TrimPrefix(raw, "-")
	}

	if strings.HasSuffix(raw, "%") {
		d.Kind = DiscountPercentage
		raw = strings.TrimSuffix(raw, "%")

		v, err := strconv.Atoi(raw)
		if err != nil {
			return d, err
		}
		d.Value = int64(v)
		return d, nil
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return d, err
	}

	d.Kind = DiscountFixed
	d.Value = int64(v)
	return d, nil
}

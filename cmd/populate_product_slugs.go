package cmd

import (
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode/sqids"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var flagsPopulateProductSlugs struct {
	dryRun bool
}

func init() {
	f := cmdPopulateProductSlugs.Flags
	f().BoolVarP(&flagsPopulateProductSlugs.dryRun, "dry-run", "d", true, "Dry run mode (don't actually update)")
	rootCmd.AddCommand(cmdPopulateProductSlugs)
}

var cmdPopulateProductSlugs = &cobra.Command{
	Use:   "populate_product_slugs",
	Short: "Populate products slugs",
	RunE: func(cmd *cobra.Command, args []string) error {
		const logtag = "[CMD POPULATE PRODUCT SLUGS]"
		db := database.New(database.DB_MODE_RW)
		defer db.Close()

		ctx := cmd.Context()

		products, err := db.GetQueries().GetProductsWithoutSlugs(ctx)
		if err != nil {
			return err
		}

		encoder := sqids.MustSqids()
		errors := make([]error, 0, len(products))
		for _, product := range products {
			id := encoder.Encode(product.ID)
			slug := utils.ProductSlug(
				product.BrandName,
				product.ProductCategory,
				product.ProductSubcategory,
				product.Serial,
				product.Power,
			)
			if flagsPopulateProductSlugs.dryRun {
				fmt.Println("product id", id, "slug", slug)
				continue
			}

			if err := db.GetQueries().UpdateProductSlugByID(ctx, queries.UpdateProductSlugByIDParams{
				ID: product.ID,
				Slug: sql.NullString{
					Valid:  true,
					String: slug,
				},
			}); err != nil {
				errors = append(errors, err)
			}
		}

		logs.Log().Info(
			logtag,
			zap.Bool("dry run", flagsPopulateProductSlugs.dryRun),
			zap.Int("products without slug", len(products)),
			zap.Int("failures", len(errors)),
			zap.Errors("errors", errors),
		)

		return nil
	},
}

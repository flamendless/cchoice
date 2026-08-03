package cmd

import (
	"cchoice/internal/cmdaudit"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var flagsPopulateBrandSlugs struct {
	dryRun bool
}

func init() {
	f := cmdPopulateBrandSlugs.Flags
	f().BoolVarP(&flagsPopulateBrandSlugs.dryRun, "dry-run", "d", true, "Dry run mode (don't actually update)")
	rootCmd.AddCommand(cmdPopulateBrandSlugs)
}

var cmdPopulateBrandSlugs = &cobra.Command{
	Use:   "populate_brand_slugs",
	Short: "Populate brand slugs using utils.BrandSlug with duplicate resolution",
	RunE: func(cmd *cobra.Command, args []string) error {
		const logtag = "[CMD POPULATE BRAND SLUGS]"
		db := database.New(database.DB_MODE_RW)
		defer db.Close()

		ctx := cmd.Context()

		brands, err := db.GetQueries().ListBrandsForSlugBackfill(ctx)
		if err != nil {
			return err
		}

		taken := make(map[string]int64, len(brands))
		updateErrors := make([]error, 0)
		updated := 0

		for _, brand := range brands {
			slugValue := utils.UniqueBrandSlug(utils.BrandSlug(brand.Name), brand.ID, taken)
			current := utils.NullStringValue(brand.Slug)
			if current == slugValue {
				continue
			}

			if flagsPopulateBrandSlugs.dryRun {
				fmt.Printf("brand id=%d name=%q slug %q -> %q\n", brand.ID, brand.Name, current, slugValue)
				continue
			}

			if err := db.GetQueries().UpdateBrandSlug(ctx, queries.UpdateBrandSlugParams{
				Slug: sql.NullString{String: slugValue, Valid: true},
				ID:   brand.ID,
			}); err != nil {
				updateErrors = append(updateErrors, err)
				continue
			}
			updated++
		}

		logs.Log().Info(
			logtag,
			zap.Bool("dry run", flagsPopulateBrandSlugs.dryRun),
			zap.Int("brands", len(brands)),
			zap.Int("updated", updated),
			zap.Int("failures", len(updateErrors)),
			zap.Errors("errors", updateErrors),
		)

		cmd.SetContext(cmdaudit.SetSummary(
			cmd.Context(),
			fmt.Sprintf("dry_run=%t brands=%d updated=%d failures=%d", flagsPopulateBrandSlugs.dryRun, len(brands), updated, len(updateErrors)),
		))

		if !flagsPopulateBrandSlugs.dryRun && len(updateErrors) == 0 {
			recordDatamigrateSuccess(cmd, args, db)
		}

		return nil
	},
}

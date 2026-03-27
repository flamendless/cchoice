package cmd

import (
	"database/sql"

	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode/sqids"
	"cchoice/internal/logs"
	"cchoice/internal/storage"
	"cchoice/internal/storage/cloudflare"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var flagsPopulateProductImagesCDN struct {
	dryRun           bool
	panicImmediately bool
	forceUpdate      bool
}

func init() {
	f := cmdPopulateProductImagesCDN.Flags
	f().BoolVarP(&flagsPopulateProductImagesCDN.dryRun, "dry-run", "d", true, "Dry run mode (don't actually update)")
	f().BoolVarP(&flagsPopulateProductImagesCDN.panicImmediately, "panic-imm", "e", true, "Panic immediately on first error")
	f().BoolVarP(&flagsPopulateProductImagesCDN.forceUpdate, "force-update", "f", false, "Force update")
	rootCmd.AddCommand(cmdPopulateProductImagesCDN)
}

var cmdPopulateProductImagesCDN = &cobra.Command{
	Use:   "populate_product_images_cdn",
	Short: "Populate CDN URLs for product images",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := conf.Conf()
		db := database.New(database.DB_MODE_RW)
		defer db.Close()

		encoder := sqids.MustSqids()

		if cfg.StorageProvider != storage.STORAGE_PROVIDER_CLOUDFLARE_IMAGES.String() {
			panic("Must use CF for this")
		}

		objectStorage, err := cloudflare.NewClientFromConfig()
		if err != nil {
			panic(err)
		}

		getCDNURL := func(path string) string {
			if path == "" {
				return path
			}
			return objectStorage.GetPublicURL(path)
		}

		ctx := cmd.Context()

		var images []queries.GetProductImagesWithEmptyCDNURLsRow
		if flagsPopulateProductImagesCDN.forceUpdate {
			res, err := db.GetQueries().GetProductImagesWithEmptyCDNURLsForce(ctx)
			if err != nil {
				panic(err)
			}
			images = make([]queries.GetProductImagesWithEmptyCDNURLsRow, 0, len(res))
			for _, d := range res {
				images = append(images, queries.GetProductImagesWithEmptyCDNURLsRow(d))
			}

		} else {
			res, err := db.GetQueries().GetProductImagesWithEmptyCDNURLs(ctx)
			if err != nil {
				panic(err)
			}
			images = make([]queries.GetProductImagesWithEmptyCDNURLsRow, 0, len(res))
			for _, d := range res {
				images = append(images, queries.GetProductImagesWithEmptyCDNURLsRow(d))
			}
		}

		totalImages := len(images)
		logs.Log().Info(
			"Found product images with empty CDN URLs",
			zap.Bool("dry_run", flagsPopulateProductImagesCDN.dryRun),
			zap.Bool("force_update", flagsPopulateProductImagesCDN.forceUpdate),
			zap.Int("total_images", totalImages),
		)

		if flagsPopulateProductImagesCDN.dryRun {
			logs.Log().Info(
				"Dry run mode - no updates will be made",
				zap.Int("images_to_process", totalImages),
			)
			return
		}

		var errorCount int
		var processed, updated int

		for _, img := range images {
			cdnURL := getCDNURL(constants.ToPath1280(img.ThumbnailPath))
			cdnURLThumbnail := getCDNURL(img.ThumbnailPath)

			logs.Log().Info(
				"Processing product image",
				zap.String("id", encoder.Encode(img.ID)),
				zap.String("product_id", encoder.Encode(img.ProductID)),
				zap.String("cdn_url", cdnURL),
				zap.String("cdn_url_thumbnail", cdnURLThumbnail),
			)

			if _, err = db.GetQueries().UpdateProductImageCDNURLs(ctx, queries.UpdateProductImageCDNURLsParams{
				ID:              img.ID,
				CdnUrl:          sql.NullString{String: cdnURL, Valid: cdnURL != ""},
				CdnUrlThumbnail: sql.NullString{String: cdnURLThumbnail, Valid: cdnURLThumbnail != ""},
			}); err != nil {
				logs.Log().Error("Failed to update product image", zap.Error(err))
				errorCount++
				if flagsPopulateProductImagesCDN.panicImmediately {
					panic(err)
				}
				continue
			}

			processed++
			updated++
		}

		logs.Log().Info(
			"CDN URL population completed",
			zap.Int("total_images", totalImages),
			zap.Int("processed", processed),
			zap.Int("updated", updated),
			zap.Int("errors", errorCount),
		)
	},
}

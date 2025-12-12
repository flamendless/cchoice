package cmd

import (
	"cchoice/internal/constants"
	"cchoice/internal/errs"
	"cchoice/internal/images"
	"cchoice/internal/logs"
	"cchoice/internal/storage/cloudflare"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var flagsMigrateImagesCloudflare struct {
	basePath         string
	dryRun           bool
	panicImmediately bool
	batchSize        int
}

func init() {
	f := cmdMigrateImagesCloudflare.Flags
	f().StringVarP(&flagsMigrateImagesCloudflare.basePath, "basepath", "p", "./cmd/web/static/images", "Base path to images directory")
	f().BoolVarP(&flagsMigrateImagesCloudflare.dryRun, "dry-run", "d", true, "Dry run mode (don't actually upload)")
	f().BoolVarP(&flagsMigrateImagesCloudflare.panicImmediately, "panic-imm", "e", true, "Panic immediately on first error")
	f().IntVarP(&flagsMigrateImagesCloudflare.batchSize, "batch-size", "b", 100, "Number of images per batch upload")
	rootCmd.AddCommand(cmdMigrateImagesCloudflare)
}

var cmdMigrateImagesCloudflare = &cobra.Command{
	Use:   "migrate_images_cloudflare",
	Short: "migrate images to Cloudflare Images",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := cloudflare.NewClientFromConfig()
		if err != nil {
			panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to initialize Cloudflare client: %w", err)))
		}

		ctx := context.Background()

		if err := client.HeadBucket(ctx); err != nil {
			panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to connect to Cloudflare Images: %w", err)))
		}

		logs.Log().Info(
			"Connected to Cloudflare Images",
			zap.String("account_id", client.GetAccountID()),
			zap.String("variant", client.GetVariant()),
			zap.Bool("dry_run", flagsMigrateImagesCloudflare.dryRun),
			zap.Int("batch_size", flagsMigrateImagesCloudflare.batchSize),
		)

		basePath := flagsMigrateImagesCloudflare.basePath

		var allImages []imageToUpload

		specificFiles := []string{
			constants.EmptyImageFilename,
			"logo.svg",
			"store.webp",
		}
		logs.Log().Info(
			"Collecting specific files from base images directory",
			zap.Strings("files", specificFiles),
		)
		for _, fileName := range specificFiles {
			filePath := filepath.Join(basePath, fileName)
			if _, err := os.Stat(filePath); err == nil {
				cfKey := "static/images/" + fileName
				allImages = append(allImages, imageToUpload{
					localPath: filePath,
					key:       cfKey,
				})
			} else {
				logs.Log().Warn(
					"Specific file not found",
					zap.String("file", filePath),
				)
			}
		}

		logosPath := filepath.Join(basePath, "logos")
		if _, err := os.Stat(logosPath); err == nil {
			logs.Log().Info(
				"Collecting logos",
				zap.String("path", logosPath),
			)
			images, err := collectImagesCloudflare(logosPath, basePath, true)
			if err != nil {
				panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to collect logos: %w", err)))
			}
			allImages = append(allImages, images...)
		} else {
			logs.Log().Warn(
				"logos directory not found",
				zap.String("path", logosPath),
			)
		}

		paymentLogosPath := filepath.Join(basePath, "payments")
		if _, err := os.Stat(paymentLogosPath); err == nil {
			logs.Log().Info(
				"Collecting payment logos",
				zap.String("path", paymentLogosPath),
			)
			images, err := collectImagesCloudflare(paymentLogosPath, basePath, true)
			if err != nil {
				panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to collect payment logos: %w", err)))
			}
			allImages = append(allImages, images...)
		} else {
			logs.Log().Warn(
				"Payment logos directory not found",
				zap.String("path", paymentLogosPath),
			)
		}

		brandLogosPath := filepath.Join(basePath, "brand_logos")
		if _, err := os.Stat(brandLogosPath); err == nil {
			logs.Log().Info(
				"Collecting brand logos",
				zap.String("path", brandLogosPath),
			)
			images, err := collectImagesCloudflare(brandLogosPath, basePath, true)
			if err != nil {
				panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to collect brand logos: %w", err)))
			}
			allImages = append(allImages, images...)
		} else {
			logs.Log().Warn(
				"Brand logos directory not found",
				zap.String("path", brandLogosPath),
			)
		}

		productImagesPath := filepath.Join(basePath, "product_images")
		if _, err := os.Stat(productImagesPath); err == nil {
			logs.Log().Info(
				"Collecting product images",
				zap.String("path", productImagesPath),
			)
			images, err := collectImagesCloudflare(productImagesPath, basePath, false)
			if err != nil {
				panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to collect product images: %w", err)))
			}
			allImages = append(allImages, images...)
		} else {
			logs.Log().Warn(
				"Product images directory not found",
				zap.String("path", productImagesPath),
			)
		}

		logs.Log().Info(
			"Total images collected",
			zap.Int("count", len(allImages)),
		)

		batchSize := flagsMigrateImagesCloudflare.batchSize
		totalBatches := (len(allImages) + batchSize - 1) / batchSize
		var totalUploaded, totalSkipped, totalErrors int

		for batchNum := range totalBatches {
			start := batchNum * batchSize
			end := min(start+batchSize, len(allImages))
			batch := allImages[start:end]

			logs.Log().Info(
				"Processing batch",
				zap.Int("batch", batchNum+1),
				zap.Int("total_batches", totalBatches),
				zap.Int("images_in_batch", len(batch)),
			)

			uploaded, skipped, errCount := processBatchCloudflare(ctx, client, batch)
			totalUploaded += uploaded
			totalSkipped += skipped
			totalErrors += errCount
		}

		logs.Log().Info(
			"Image migration completed",
			zap.Int("total_uploaded", totalUploaded),
			zap.Int("total_skipped", totalSkipped),
			zap.Int("total_errors", totalErrors),
		)
	},
}

type imageToUpload struct {
	localPath string
	key       string
}

func collectImagesCloudflare(imagesPath string, basePath string, isBrandLogos bool) ([]imageToUpload, error) {
	var imgsToUpload []imageToUpload

	err := filepath.Walk(imagesPath, func(filePath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(filePath))
		imgFormat := images.ParseImageFormatExtToEnum(ext)
		if imgFormat == images.IMAGE_FORMAT_UNDEFINED {
			return nil
		}

		if isBrandLogos && imgFormat == images.IMAGE_FORMAT_PNG {
			logs.Log().Debug(
				"Skipping PNG file for brand logos (only WebP allowed)",
				zap.String("file", filePath),
			)
			return nil
		}

		relPath, err := filepath.Rel(basePath, filePath)
		if err != nil {
			logs.Log().Error(
				"Failed to calculate relative path",
				zap.Error(err),
				zap.String("file", filePath),
			)
			if flagsMigrateImagesCloudflare.panicImmediately {
				panic(err)
			}
			return nil
		}

		normalizedPath := strings.ReplaceAll(relPath, string(filepath.Separator), "/")
		if strings.Contains(normalizedPath, "/original/") {
			logs.Log().Debug(
				"Skipping original image",
				zap.String("file", filePath),
				zap.String("rel_path", relPath),
			)
			return nil
		}

		cfKey := "static/images/" + normalizedPath
		imgsToUpload = append(imgsToUpload, imageToUpload{
			localPath: filePath,
			key:       cfKey,
		})

		return nil
	})

	return imgsToUpload, err
}

func processBatchCloudflare(ctx context.Context, client *cloudflare.Client, batch []imageToUpload) (uploaded, skipped, errCount int) {
	for _, img := range batch {
		info, err := os.Stat(img.localPath)
		if err != nil {
			logs.Log().Error(
				"Failed to stat file",
				zap.Error(err),
				zap.String("file", img.localPath),
			)
			errCount++
			if flagsMigrateImagesCloudflare.panicImmediately {
				panic(err)
			}
			continue
		}

		if info.IsDir() {
			continue
		}

		if !flagsMigrateImagesCloudflare.dryRun {
			exists, err := client.ObjectExists(ctx, img.key)
			if err == nil && exists {
				logs.Log().Debug(
					"Image already exists, skipping",
					zap.String("key", img.key),
				)
				skipped++
				continue
			}
		}

		if flagsMigrateImagesCloudflare.dryRun {
			fmt.Printf("Would upload %s -> %s\n", img.localPath, img.key)
			uploaded++
			continue
		}

		logs.Log().Info(
			"Uploading file",
			zap.String("local", img.localPath),
			zap.String("key", img.key),
			zap.Int64("size", info.Size()),
		)

		data, err := os.ReadFile(img.localPath)
		if err != nil {
			logs.Log().Error(
				"Failed to read file",
				zap.Error(err),
				zap.String("file", img.localPath),
			)
			errCount++
			if flagsMigrateImagesCloudflare.panicImmediately {
				panic(err)
			}
			continue
		}

		ext := strings.ToLower(filepath.Ext(img.localPath))
		imgFormat := images.ParseImageFormatExtToEnum(ext)
		contentType := imgFormat.MIMEType()

		if err := client.PutObjectFromBytes(ctx, img.key, data, contentType); err != nil {
			logs.Log().Error(
				"Failed to upload file",
				zap.Error(err),
				zap.String("file", img.localPath),
				zap.String("key", img.key),
			)
			errCount++
			if flagsMigrateImagesCloudflare.panicImmediately {
				panic(err)
			}
			continue
		}

		logs.Log().Debug(
			"Uploaded",
			zap.String("local", img.localPath),
			zap.String("key", img.key),
			zap.String("content_type", contentType),
			zap.Int64("size", info.Size()),
		)
		uploaded++
	}

	return uploaded, skipped, errCount
}

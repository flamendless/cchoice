package cmd

import (
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/storage/linode"
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

var flagsMigrateImagesLinode struct {
	basePath         string
	dryRun           bool
	panicImmediately bool
	bucket           string
}

func init() {
	f := cmdMigrateImagesLinode.Flags
	f().StringVarP(&flagsMigrateImagesLinode.basePath, "basepath", "p", "./cmd/web/static/images", "Base path to images directory")
	f().BoolVarP(&flagsMigrateImagesLinode.dryRun, "dry-run", "d", true, "Dry run mode (don't actually upload)")
	f().BoolVarP(&flagsMigrateImagesLinode.panicImmediately, "panic-imm", "e", true, "Panic immediately on first error")
	f().StringVarP(&flagsMigrateImagesLinode.bucket, "bucket", "b", "PRIVATE", "Bucket enum to use (PUBLIC or PRIVATE)")
	rootCmd.AddCommand(cmdMigrateImagesLinode)
}

var cmdMigrateImagesLinode = &cobra.Command{
	Use:   "migrate_images_linode",
	Short: "migrate product images and brand logos to Linode Object Storage",
	Run: func(cmd *cobra.Command, args []string) {
		bucketEnum := enums.ParseLinodeBucketEnum(strings.ToUpper(flagsMigrateImagesLinode.bucket))
		if bucketEnum == enums.LINODE_BUCKET_UNDEFINED {
			panic(errors.Join(errs.ErrCmd, fmt.Errorf("invalid bucket enum: %s (valid values: PUBLIC or PRIVATE)", flagsMigrateImagesLinode.bucket)))
		}

		client, err := linode.NewClientFromConfigWithBucket(bucketEnum)
		if err != nil {
			panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to initialize Linode client: %w", err)))
		}

		ctx := context.Background()

		if err := client.HeadBucket(ctx); err != nil {
			panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to connect to Linode bucket: %w", err)))
		}

		logs.Log().Info(
			"Connected to Linode",
			zap.String("bucket", client.GetBucket()),
			zap.String("bucket_enum", bucketEnum.String()),
			zap.String("prefix", client.GetBasePrefix()),
			zap.Bool("dry_run", flagsMigrateImagesLinode.dryRun),
		)

		basePath := flagsMigrateImagesLinode.basePath

		specificFiles := []string{"empty_96x96.webp", "logo.svg", "store.webp"}
		logs.Log().Info(
			"Migrating specific files from base images directory",
			zap.Strings("files", specificFiles),
		)
		for _, fileName := range specificFiles {
			filePath := filepath.Join(basePath, fileName)
			if _, err := os.Stat(filePath); err == nil {
				s3Key := "static/images/" + fileName
				if err := migrateFile(ctx, client, filePath, s3Key); err != nil {
					logs.Log().Error(
						"Failed to migrate specific file",
						zap.Error(err),
						zap.String("file", filePath),
					)
					if flagsMigrateImagesLinode.panicImmediately {
						panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to migrate file %s: %w", fileName, err)))
					}
				}
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
				"Migrating logos",
				zap.String("path", logosPath),
			)
			if err := migrateImages(ctx, client, logosPath, basePath, "logos", true); err != nil {
				panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to migrate logos: %w", err)))
			}
		} else {
			logs.Log().Warn(
				"logos directory not found",
				zap.String("path", logosPath),
			)
		}

		paymentLogosPath := filepath.Join(basePath, "payments")
		if _, err := os.Stat(paymentLogosPath); err == nil {
			logs.Log().Info(
				"Migrating payment logos",
				zap.String("path", paymentLogosPath),
			)
			if err := migrateImages(ctx, client, paymentLogosPath, basePath, "Payment logos", true); err != nil {
				panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to migrate payment logos: %w", err)))
			}
		} else {
			logs.Log().Warn(
				"Payment logos directory not found",
				zap.String("path", paymentLogosPath),
			)
		}

		brandLogosPath := filepath.Join(basePath, "brand_logos")
		if _, err := os.Stat(brandLogosPath); err == nil {
			logs.Log().Info(
				"Migrating brand logos",
				zap.String("path", brandLogosPath),
			)
			if err := migrateImages(ctx, client, brandLogosPath, basePath, "Brand logos", true); err != nil {
				panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to migrate brand logos: %w", err)))
			}
		} else {
			logs.Log().Warn(
				"Brand logos directory not found",
				zap.String("path", brandLogosPath),
			)
		}

		productImagesPath := filepath.Join(basePath, "product_images")
		if _, err := os.Stat(productImagesPath); err == nil {
			logs.Log().Info(
				"Migrating product images",
				zap.String("path", productImagesPath),
			)
			if err := migrateImages(ctx, client, productImagesPath, basePath, "Product images", false); err != nil {
				panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to migrate product images: %w", err)))
			}
		} else {
			logs.Log().Warn(
				"Product images directory not found",
				zap.String("path", productImagesPath),
			)
		}

		logs.Log().Info("Image migration completed")
	},
}

func migrateImages(
	ctx context.Context,
	client *linode.Client,
	imagesPath string,
	basePath string,
	summaryLabel string,
	isBrandLogos bool,
) error {
	var totalFiles int
	var uploadedFiles int
	var skippedFiles int
	var errorFiles int

	err := filepath.Walk(imagesPath, func(filePath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(filePath))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" && ext != ".gif" && ext != ".svg" {
			return nil
		}

		if isBrandLogos && ext == ".png" {
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
			errorFiles++
			if flagsMigrateImagesLinode.panicImmediately {
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

		totalFiles++

		s3Key := "static/images/" + normalizedPath
		normalizedKey := strings.TrimPrefix(s3Key, "static/")
		if !flagsMigrateImagesLinode.dryRun {
			exists, err := client.ObjectExists(ctx, s3Key)
			if err == nil && exists {
				skippedFiles++
				_ = migrateFile(ctx, client, filePath, s3Key)
				return nil
			}
		}

		if err := migrateFile(ctx, client, filePath, s3Key); err != nil {
			logs.Log().Error(
				"Failed to migrate file",
				zap.Error(err),
				zap.String("file", filePath),
				zap.String("s3_key", normalizedKey),
			)
			errorFiles++
			if flagsMigrateImagesLinode.panicImmediately {
				panic(err)
			}
			return nil
		}

		uploadedFiles++

		return nil
	})

	if err != nil {
		if flagsMigrateImagesLinode.panicImmediately {
			panic(err)
		}
		return err
	}

	logs.Log().Info(
		summaryLabel+" migration summary",
		zap.Int("total", totalFiles),
		zap.Int("uploaded", uploadedFiles),
		zap.Int("skipped", skippedFiles),
		zap.Int("errors", errorFiles),
	)

	return nil
}

func getContentType(ext string) string {
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".md":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	case ".html":
		return "text/html"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

func migrateFile(ctx context.Context, client *linode.Client, filePath string, s3Key string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("file path is a directory, not a file: %s", filePath)
	}

	normalizedKey := strings.TrimPrefix(s3Key, "static/")
	if !flagsMigrateImagesLinode.dryRun {
		exists, err := client.ObjectExists(ctx, s3Key)
		if err == nil && exists {
			logs.Log().Debug(
				"Object already exists, skipping",
				zap.String("s3_key", normalizedKey),
			)
			return nil
		}
	}

	if flagsMigrateImagesLinode.dryRun {
		fmt.Printf("Would upload %s -> {%s}/%s\n", filePath, flagsMigrateImagesLinode.bucket, normalizedKey)
		return nil
	}

	logs.Log().Info(
		"Uploading file",
		zap.String("local", filePath),
		zap.String("s3_key", normalizedKey),
		zap.Int64("size", info.Size()),
	)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	contentType := getContentType(ext)

	if err := client.PutObjectFromBytes(ctx, s3Key, data, contentType); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	logs.Log().Debug(
		"Uploaded",
		zap.String("local", filePath),
		zap.String("s3_key", normalizedKey),
		zap.String("content_type", contentType),
		zap.Int64("size", info.Size()),
	)

	return nil
}

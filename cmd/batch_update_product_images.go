package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	"cchoice/internal/cmdaudit"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/services"
	"cchoice/internal/storage"
	"cchoice/internal/storage/cloudflare"
	"cchoice/internal/utils"
)

const batchUpdateProductImagesLogtag = "[BUPI]"

type batchImageFile struct {
	path        string
	filename    string
	contentType string
}

type batchProductRow struct {
	queries.GetProductsWithImagesByBrandIDRow
}

type batchUpdateStats struct {
	total     int
	matched   int
	update    int
	uploaded  int
	retain    int
	noImage   int
	noProduct int
	errCount  int
}

var flagsBatchUpdateProductImages struct {
	dryRun     bool
	input      string
	imagesDir  string
	codeColumn string
	brand      string
	onlyBefore string
}

var cmdBatchUpdateProductImages = &cobra.Command{
	Use:   "batch_update_product_images",
	Short: "Batch update product images from a spreadsheet and local image folder",
	RunE:  runBatchUpdateProductImages,
}

func init() {
	f := cmdBatchUpdateProductImages.Flags
	f().BoolVarP(&flagsBatchUpdateProductImages.dryRun, "dry-run", "d", true, "Dry run mode (don't actually upload or update)")
	f().StringVarP(&flagsBatchUpdateProductImages.input, "input", "i", "", "Input spreadsheet path (.csv or .xlsx)")
	f().StringVarP(&flagsBatchUpdateProductImages.imagesDir, "images-dir", "p", "", "Folder containing product images")
	f().StringVarP(&flagsBatchUpdateProductImages.codeColumn, "code-column", "c", "", "Spreadsheet column header for product codes")
	f().StringVarP(&flagsBatchUpdateProductImages.brand, "brand", "b", "", "Brand name (scopes DB lookup and strips prefix during code matching)")
	f().StringVar(&flagsBatchUpdateProductImages.onlyBefore, "only-before", "", "Cutoff datetime; images at or after this are retained (default: start of today PH)")
	if err := cmdBatchUpdateProductImages.MarkFlagRequired("input"); err != nil {
		panic(err)
	}
	if err := cmdBatchUpdateProductImages.MarkFlagRequired("images-dir"); err != nil {
		panic(err)
	}
	if err := cmdBatchUpdateProductImages.MarkFlagRequired("code-column"); err != nil {
		panic(err)
	}
	if err := cmdBatchUpdateProductImages.MarkFlagRequired("brand"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(cmdBatchUpdateProductImages)
}

func runBatchUpdateProductImages(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if conf.Conf().StorageProvider != storage.STORAGE_PROVIDER_CLOUDFLARE_IMAGES.String() {
		return fmt.Errorf("%w: storage provider must be Cloudflare Images", errs.ErrCmd)
	}

	cutoff, err := utils.ParseOnlyBeforePH(flagsBatchUpdateProductImages.onlyBefore)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrCmdInvalidFlag, err)
	}

	codes, err := parseBatchImageCodes(flagsBatchUpdateProductImages.input, flagsBatchUpdateProductImages.codeColumn)
	if err != nil {
		return err
	}

	imageIndex, err := buildBatchImageIndex(flagsBatchUpdateProductImages.imagesDir, flagsBatchUpdateProductImages.brand)
	if err != nil {
		return err
	}

	db := database.New(database.DB_MODE_RW)
	defer db.Close()

	brandID, err := db.GetQueries().GetBrandsIDByName(ctx, flagsBatchUpdateProductImages.brand)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: brand %q does not exist", errs.ErrCmdInvalidFlag, flagsBatchUpdateProductImages.brand)
	}
	if err != nil {
		return err
	}

	objectStorage, err := cloudflare.NewClientFromConfig()
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrCmd, err)
	}
	if err := objectStorage.HeadBucket(ctx); err != nil {
		return fmt.Errorf("%w: failed to connect to Cloudflare Images: %w", errs.ErrCmd, err)
	}

	productRows, err := db.GetQueries().GetProductsWithImagesByBrandID(ctx, brandID)
	if err != nil {
		return err
	}
	productIndex := buildBatchProductIndex(productRows, flagsBatchUpdateProductImages.brand)

	imageSvc := services.NewImageService(objectStorage, nil, nil, nil)
	thumbSvc := services.NewThumbnailService(objectStorage)

	stats := batchUpdateStats{}

	for _, rawCode := range codes {
		stats.total++
		normalized := utils.NormalizeProductCode(rawCode, flagsBatchUpdateProductImages.brand)

		img, ok := imageIndex[normalized]
		if !ok {
			logs.Log().Info(
				batchUpdateProductImagesLogtag,
				zap.String("product_code", rawCode),
				zap.String("outcome", "found no matching image in folder"),
			)
			stats.noImage++
			continue
		}

		product, ok := productIndex[normalized]
		if !ok {
			logs.Log().Info(
				batchUpdateProductImagesLogtag,
				zap.String("product_code", rawCode),
				zap.String("outcome", "found no matching product in database"),
			)
			stats.noProduct++
			continue
		}

		stats.matched++

		if product.ProductImageID.Valid && product.ProductImageID.Int64 > 0 {
			shouldRetain, latest := shouldRetainProductImage(product.GetProductsWithImagesByBrandIDRow, cutoff)
			if shouldRetain {
				logs.Log().Info(
					batchUpdateProductImagesLogtag,
					zap.String("product_code", rawCode),
					zap.String("outcome", "will retain image"),
					zap.String("image_timestamp", utils.FormatTimePH(latest)),
				)
				stats.retain++
				continue
			}
		}

		logs.Log().Info(
			batchUpdateProductImagesLogtag,
			zap.String("product_code", rawCode),
			zap.String("outcome", "will update image"),
		)
		stats.update++

		if flagsBatchUpdateProductImages.dryRun {
			continue
		}

		if err := uploadAndUpdateProductImage(
			ctx,
			db,
			imageSvc,
			thumbSvc,
			objectStorage,
			flagsBatchUpdateProductImages.brand,
			rawCode,
			product,
			img,
		); err != nil {
			logs.Log().Error(
				batchUpdateProductImagesLogtag,
				zap.String("code", rawCode),
				zap.Error(err),
			)
			stats.errCount++
			continue
		}
		stats.uploaded++
	}

	logs.Log().Info(
		batchUpdateProductImagesLogtag,
		zap.String("result", "completed"),
		zap.Bool("dry_run", flagsBatchUpdateProductImages.dryRun),
		zap.String("brand", flagsBatchUpdateProductImages.brand),
		zap.Int("total", stats.total),
		zap.Int("matched", stats.matched),
		zap.Int("no_image", stats.noImage),
		zap.Int("no_product", stats.noProduct),
		zap.Int("retain", stats.retain),
		zap.Int("will_update", stats.update),
		zap.Int("uploaded", stats.uploaded),
		zap.Int("errors", stats.errCount),
	)

	cmd.SetContext(cmdaudit.SetSummary(
		cmd.Context(),
		fmt.Sprintf(
			"dry_run=%t brand=%s total=%d matched=%d update=%d uploaded=%d retain=%d no_image=%d no_product=%d errors=%d",
			flagsBatchUpdateProductImages.dryRun,
			flagsBatchUpdateProductImages.brand,
			stats.total,
			stats.matched,
			stats.update,
			stats.uploaded,
			stats.retain,
			stats.noImage,
			stats.noProduct,
			stats.errCount,
		),
	))

	return nil
}

func parseBatchImageCodes(path, codeColumn string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: input file not found: %s", errs.ErrCmdInvalidFlag, path)
		}
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%w: input path is a directory: %s", errs.ErrCmdInvalidFlag, path)
	}

	format := enums.ParseOutputFormatExtToEnum(filepath.Ext(path))
	if format == enums.OUTPUT_FORMAT_UNDEFINED {
		return nil, fmt.Errorf("%w: unsupported input file type: %s", errs.ErrCmdInvalidFlag, filepath.Ext(path))
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	headers, records, err := parseBatchImageSpreadsheet(format, f)
	if err != nil {
		return nil, err
	}

	colIndex, err := findBatchImageColumnIndex(headers, codeColumn)
	if err != nil {
		return nil, err
	}

	codes := make([]string, 0, len(records))
	for _, row := range records {
		if colIndex >= len(row) {
			continue
		}
		code := strings.TrimSpace(row[colIndex])
		if code == "" {
			continue
		}
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return nil, fmt.Errorf("%w: spreadsheet has no product codes", errs.ErrCmdInvalidFlag)
	}

	return codes, nil
}

func parseBatchImageSpreadsheet(format enums.OutputFormat, r io.Reader) ([]string, [][]string, error) {
	switch format {
	case enums.OUTPUT_FORMAT_CSV:
		return parseBatchImageCSV(r)
	case enums.OUTPUT_FORMAT_XLSX:
		return parseBatchImageXLSX(r)
	default:
		return nil, nil, fmt.Errorf("%w: unsupported file type: %s", errs.ErrCmdInvalidFlag, format.Extension())
	}
}

func parseBatchImageCSV(r io.Reader) ([]string, [][]string, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to read csv: %w", errs.ErrCmdInvalidFlag, err)
	}
	if len(records) < 2 {
		return nil, nil, fmt.Errorf("%w: csv has no data rows", errs.ErrCmdInvalidFlag)
	}

	return records[0], records[1:], nil
}

func parseBatchImageXLSX(r io.Reader) ([]string, [][]string, error) {
	file, err := excelize.OpenReader(r)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to read xlsx: %w", errs.ErrCmdInvalidFlag, err)
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, fmt.Errorf("%w: xlsx has no sheets", errs.ErrCmdInvalidFlag)
	}

	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to read xlsx rows: %w", errs.ErrCmdInvalidFlag, err)
	}
	if len(rows) < 2 {
		return nil, nil, fmt.Errorf("%w: xlsx has no data rows", errs.ErrCmdInvalidFlag)
	}

	headers := rows[0]
	records := make([][]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		if batchImageEmptyRow(row) {
			continue
		}
		values := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				values[i] = strings.TrimSpace(row[i])
			}
		}
		records = append(records, values)
	}

	return headers, records, nil
}

func batchImageEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func findBatchImageColumnIndex(headers []string, codeColumn string) (int, error) {
	target := strings.ToLower(strings.TrimSpace(codeColumn))
	for i, h := range headers {
		if strings.ToLower(strings.TrimSpace(h)) == target {
			return i, nil
		}
	}
	return -1, fmt.Errorf("%w: column %q not found in spreadsheet", errs.ErrCmdMissingColumn, codeColumn)
}

func buildBatchImageIndex(imagesDir, brand string) (map[string]batchImageFile, error) {
	info, err := os.Stat(imagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: images directory not found: %s", errs.ErrCmdInvalidFlag, imagesDir)
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: images path is not a directory: %s", errs.ErrCmdInvalidFlag, imagesDir)
	}

	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		return nil, err
	}

	index := make(map[string]batchImageFile)
	validCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if !enums.IsValidImageExtension(ext) {
			continue
		}

		path := filepath.Join(imagesDir, name)
		contentType, err := readBatchImageContentType(path)
		if err != nil {
			logs.Log().Warn(
				batchUpdateProductImagesLogtag,
				zap.String("file", path),
				zap.Error(err),
			)
			continue
		}

		validCount++
		key := utils.NormalizeProductCode(strings.TrimSuffix(name, ext), brand)
		if _, exists := index[key]; exists {
			logs.Log().Warn(
				batchUpdateProductImagesLogtag,
				zap.String("key", key),
				zap.String("file", path),
				zap.String("reason", "duplicate normalized code, keeping first match"),
			)
			continue
		}

		index[key] = batchImageFile{
			path:        path,
			filename:    name,
			contentType: contentType,
		}
	}

	if validCount == 0 {
		return nil, fmt.Errorf("%w: images directory has no valid image files: %s", errs.ErrCmdInvalidFlag, imagesDir)
	}

	return index, nil
}

func readBatchImageContentType(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	return utils.DetectProductImageContentType(buf[:n])
}

func buildBatchProductIndex(rows []queries.GetProductsWithImagesByBrandIDRow, brand string) map[string]batchProductRow {
	index := make(map[string]batchProductRow, len(rows))
	for _, row := range rows {
		key := utils.NormalizeProductCode(row.Serial, brand)
		if _, exists := index[key]; exists {
			logs.Log().Warn(
				batchUpdateProductImagesLogtag,
				zap.String("key", key),
				zap.String("serial", row.Serial),
				zap.String("reason", "duplicate normalized serial, keeping first match"),
			)
			continue
		}
		index[key] = batchProductRow{GetProductsWithImagesByBrandIDRow: row}
	}
	return index
}

func shouldRetainProductImage(row queries.GetProductsWithImagesByBrandIDRow, cutoff time.Time) (bool, time.Time) {
	var latest time.Time
	if row.ImageCreatedAt.Valid {
		latest = row.ImageCreatedAt.Time
	}
	if row.ImageUpdatedAt.Valid && row.ImageUpdatedAt.Time.After(latest) {
		latest = row.ImageUpdatedAt.Time
	}
	if latest.IsZero() {
		return false, time.Time{}
	}
	return !latest.Before(cutoff), latest
}

func uploadAndUpdateProductImage(
	ctx context.Context,
	db database.IService,
	imageSvc *services.ImageService,
	thumbSvc *services.ThumbnailService,
	objectStorage storage.IObjectStorage,
	brandName, rawCode string,
	product batchProductRow,
	img batchImageFile,
) error {
	data, err := os.ReadFile(img.path)
	if err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(img.filename))
	generatedKey := imageSvc.GenerateFilename(
		enums.IMAGE_PREFIX_PRODUCT_IMAGE,
		ext,
		brandName,
		product.Name,
	)

	logs.Log().Info(
		batchUpdateProductImagesLogtag,
		zap.String("action", "uploading"),
		zap.String("filename", img.filename),
		zap.String("cloudflare_key", generatedKey),
		zap.String("product_code", rawCode),
	)

	if err := imageSvc.UploadProductImage(
		ctx,
		brandName,
		generatedKey,
		bytes.NewReader(data),
		img.contentType,
	); err != nil {
		return err
	}

	variants, err := thumbSvc.ProcessImageVariants(ctx, generatedKey, brandName, generatedKey)
	if err != nil {
		return err
	}

	thumbnailURL := generatedKey
	for _, v := range variants {
		if v.Size == "640x640" {
			thumbnailURL = v.URL
			break
		}
	}

	if conf.Conf().IsLocal() {
		thumbnailURL = "static/" + thumbnailURL
	}

	cdnURL := objectStorage.GetPublicURL(generatedKey)
	cdnURLThumbnail := objectStorage.GetPublicURL(constants.ToPath1280(thumbnailURL))

	if product.ProductImageID.Valid && product.ProductImageID.Int64 > 0 {
		if err := db.GetQueries().UpdateProductImage(ctx, queries.UpdateProductImageParams{
			ID:              product.ProductImageID.Int64,
			Path:            generatedKey,
			Thumbnail:       sql.NullString{String: thumbnailURL, Valid: thumbnailURL != ""},
			CdnUrl:          sql.NullString{String: cdnURL, Valid: cdnURL != ""},
			CdnUrlThumbnail: sql.NullString{String: cdnURLThumbnail, Valid: cdnURLThumbnail != ""},
		}); err != nil {
			return err
		}
		return nil
	}

	_, err = db.GetQueries().CreateProductImage(ctx, queries.CreateProductImageParams{
		ProductID:       product.ID,
		Path:            generatedKey,
		Thumbnail:       sql.NullString{String: thumbnailURL, Valid: thumbnailURL != ""},
		CdnUrl:          sql.NullString{String: cdnURL, Valid: cdnURL != ""},
		CdnUrlThumbnail: sql.NullString{String: cdnURLThumbnail, Valid: cdnURLThumbnail != ""},
	})
	return err
}

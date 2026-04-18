package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"cchoice/internal/jobs"
	"cchoice/internal/logs"
	"cchoice/internal/storage"
	"cchoice/internal/types"
	"cchoice/internal/utils"

	"github.com/davidbyttow/govips/v2/vips"
	"go.uber.org/zap"
)

const (
	LocalImageBasePath = "cmd/web/static/images/product_images"
)

type ThumbnailService struct {
	objectStorage storage.IObjectStorage
}

func NewThumbnailService(objectStorage storage.IObjectStorage) *ThumbnailService {
	if objectStorage == nil || reflect.ValueOf(objectStorage).IsNil() {
		panic("implementor of IObjectStorage is required")
	}
	return &ThumbnailService{
		objectStorage: objectStorage,
	}
}

func (s *ThumbnailService) ProcessImageVariants(ctx context.Context, sourcePath, brand, filename string) ([]types.ThumbnailVariant, error) {
	const logtag = "[ThumbnailService ProcessImageVariants]"

	sourceExt := filepath.Ext(filename)
	sourceName := strings.TrimSuffix(filename, sourceExt)

	var sourceImageData []byte
	var err error

	isLocalStorage := s.objectStorage.ProviderEnum() == storage.STORAGE_PROVIDER_LOCAL
	if !conf.Conf().Test.LocalUploadImage || isLocalStorage {
		localPath := filepath.Join("cmd/web/static/images/product_images", strings.ToLower(brand), "original", filepath.Base(filename))
		sourceImageData, err = os.ReadFile(localPath)
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.String("brand", brand),
				zap.String("sourcePath", sourcePath),
				zap.String("filename", filename),
				zap.String("sourceName", sourceName),
				zap.String("path", localPath),
				zap.Error(err),
			)
			return nil, errors.Join(errs.ErrFS, err)
		}
	} else {
		sourceImageData, err = s.objectStorage.GetObjectBytes(ctx, sourcePath)
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.String("brand", brand),
				zap.String("path", sourcePath),
				zap.Error(err),
			)
			return nil, errors.Join(errs.ErrFS, err)
		}
	}

	img, err := vips.NewImageFromBuffer(sourceImageData)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return nil, fmt.Errorf("failed to load image: %w", err)
	}
	defer img.Close()

	sourceWidth := img.Width()
	sourceHeight := img.Height()

	logs.LogCtx(ctx).Info(logtag,
		zap.String("brand", brand),
		zap.String("filename", filename),
		zap.Int("source_width", sourceWidth),
		zap.Int("source_height", sourceHeight),
	)

	webpExport := vips.NewDefaultWEBPExportParams()

	variants := make([]types.ThumbnailVariant, 0, len(types.ImageSizes)+1)

	originalKey := s.buildStorageKey(brand, "original", sourceName, sourceExt, isLocalStorage)
	if isLocalStorage {
		originalPath := filepath.Join(LocalImageBasePath, brand, "original", filename)
		if err := os.WriteFile(originalPath, sourceImageData, 0644); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.String("path", originalPath), zap.Error(err))
			return nil, errors.Join(errs.ErrFS, err)
		}
		originalURL := fmt.Sprintf("static/images/product_images/%s/original/%s", brand, filename)
		variants = append(variants, types.ThumbnailVariant{
			Size:       "original",
			Path:       originalPath,
			URL:        originalURL,
			IsOriginal: true,
		})
		logs.LogCtx(ctx).Info(logtag, zap.String("stored", "local"), zap.String("path", originalPath))
	} else {
		if err := s.objectStorage.PutObjectFromBytes(ctx, originalKey, sourceImageData, utils.GetContentType(sourceExt)); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.String("key", originalKey), zap.Error(err))
			return nil, err
		}
		originalURL := s.objectStorage.GetPublicURL(originalKey)
		variants = append(variants, types.ThumbnailVariant{
			Size:       "original",
			Path:       originalKey,
			URL:        originalURL,
			IsOriginal: true,
		})
		logs.LogCtx(ctx).Info(logtag, zap.String("stored", "cloud"), zap.String("key", originalKey), zap.String("url", originalURL))
	}

	for _, size := range types.ImageSizes {
		folderName := fmt.Sprintf("%dx%d", size.Width, size.Height)

		if sourceWidth < size.Width || sourceHeight < size.Height {
			logs.LogCtx(ctx).Info(logtag,
				zap.String("size", folderName),
				zap.String("reason", "source smaller than target, using source for both 640x640 and 1280x1280"),
			)
			variant := types.ThumbnailVariant{
				Size:       folderName,
				Path:       sourcePath,
				URL:        sourcePath,
				IsOriginal: true,
			}
			variants = append(variants, variant)
			continue
		}

		thumbnailImg, err := vips.NewImageFromBuffer(sourceImageData)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			return nil, fmt.Errorf("failed to load image for thumbnail: %w", err)
		}
		defer thumbnailImg.Close()

		if err := thumbnailImg.Thumbnail(size.Width, size.Height, vips.InterestingCentre); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			return nil, fmt.Errorf("failed to create thumbnail: %w", err)
		}

		imgBytes, _, err := thumbnailImg.Export(webpExport)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			return nil, fmt.Errorf("failed to export WebP: %w", err)
		}

		webpFilename := sourceName + ".webp"
		storageKey := s.buildStorageKey(brand, folderName, sourceName, ".webp", isLocalStorage)

		if isLocalStorage {
			webpPath := filepath.Join(LocalImageBasePath, brand, "webp", folderName, webpFilename)
			if err := os.WriteFile(webpPath, imgBytes, 0644); err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.String("path", webpPath), zap.Error(err))
				return nil, errors.Join(errs.ErrFS, err)
			}
			webpURL := fmt.Sprintf("static/images/product_images/%s/webp/%s/%s", brand, folderName, webpFilename)
			variants = append(variants, types.ThumbnailVariant{
				Size: folderName,
				Path: webpPath,
				URL:  webpURL,
			})
			logs.LogCtx(ctx).Info(logtag, zap.String("stored", "local"), zap.String("size", folderName), zap.String("path", webpPath))
		} else {
			if err := s.objectStorage.PutObjectFromBytes(ctx, storageKey, imgBytes, "image/webp"); err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.String("key", storageKey), zap.Error(err))
				return nil, err
			}
			webpURL := s.objectStorage.GetPublicURL(storageKey)
			variants = append(variants, types.ThumbnailVariant{
				Size: folderName,
				Path: storageKey,
				URL:  webpURL,
			})
			logs.LogCtx(ctx).Info(logtag, zap.String("stored", "cloud"), zap.String("size", folderName), zap.String("key", storageKey), zap.String("url", webpURL))
		}
	}

	return variants, nil
}

func (s *ThumbnailService) buildStorageKey(brand, sizeFolder, filename, ext string, isLocal bool) string {
	if isLocal {
		return filepath.Join(LocalImageBasePath, brand, sizeFolder, filename+ext)
	}
	return fmt.Sprintf("product_images/%s/%s/%s%s", brand, sizeFolder, filename, ext)
}

func (s *ThumbnailService) ID() string {
	return "Thumbnail"
}

func (s *ThumbnailService) Log() {
	logs.Log().Info("[ThumbnailService] Loaded")
}

var _ IService = (*ThumbnailService)(nil)
var _ jobs.IThumbnailService = (*ThumbnailService)(nil)

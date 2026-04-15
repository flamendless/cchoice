package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cchoice/internal/conf"
	"cchoice/internal/database"
	"cchoice/internal/encode"
	"cchoice/internal/logs"
	"cchoice/internal/storage"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

type ImageService struct {
	objectStorage storage.IObjectStorage
	encoder       encode.IEncode
	dbRO          database.IService
	dbRW          database.IService
}

func NewImageService(
	objectStorage storage.IObjectStorage,
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
) *ImageService {
	return &ImageService{
		objectStorage: objectStorage,
		encoder:       encoder,
		dbRO:          dbRO,
		dbRW:          dbRW,
	}
}

func (s *ImageService) GenerateFilename(
	ext string,
	paths ...string,
) string {
	uuid := utils.GenString(16)
	sanitizedName := strings.ReplaceAll(strings.Join(paths, "_"), " ", "_")
	var id string
	if !conf.Conf().IsProd() {
		id = "DEV_"
	}
	return fmt.Sprintf("%s%s_%s%s", id, sanitizedName, uuid, ext)
}

func (s *ImageService) ValidateContentType(contentType string) error {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	if !validTypes[contentType] {
		return fmt.Errorf("invalid content type: %s", contentType)
	}
	return nil
}

func (s *ImageService) ValidateSize(file io.Reader) ([]byte, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(data) > 10*1024*1024 {
		return nil, fmt.Errorf("file too large: %d bytes", len(data))
	}
	return data, nil
}

func (s *ImageService) UploadProductImage(
	ctx context.Context,
	brand string,
	filename string,
	file io.Reader,
	contentType string,
) error {
	const logtag = "[ImageService] Product"
	if err := s.ValidateContentType(contentType); err != nil {
		return err
	}
	data, err := s.ValidateSize(file)
	if err != nil {
		return err
	}

	file = bytes.NewReader(data)
	isLocalStorage := s.objectStorage.ProviderEnum() == storage.STORAGE_PROVIDER_LOCAL

	logs.Log().Info(
		logtag,
		zap.String("storing product image", filename),
		zap.Stringer("using", s.objectStorage.ProviderEnum()),
		zap.String("brand", brand),
		zap.Bool("local storage", isLocalStorage),
	)

	if !conf.Conf().Test.LocalUploadImage || isLocalStorage {
		sourceName := filepath.Base(filename)
		localPath := filepath.Join("cmd/web/static/images/product_images", strings.ToLower(brand), "original", sourceName)
		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			return err
		}
		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		if err := os.WriteFile(localPath, data, 0644); err != nil {
			return err
		}
		return nil
	}

	return s.objectStorage.PutObject(ctx, filename, file, contentType)
}

func (s *ImageService) UploadBrandImage(
	ctx context.Context,
	brand string,
	filename string,
	file io.Reader,
	contentType string,
) (string, error) {
	const logtag = "[ImageService] Brand"
	if err := s.ValidateContentType(contentType); err != nil {
		return "", err
	}
	data, err := s.ValidateSize(file)
	if err != nil {
		return "", err
	}

	file = bytes.NewReader(data)
	isLocalStorage := s.objectStorage.ProviderEnum() == storage.STORAGE_PROVIDER_LOCAL

	logs.Log().Info(
		logtag,
		zap.String("storing brand image", filename),
		zap.Stringer("using", s.objectStorage.ProviderEnum()),
		zap.String("brand", brand),
		zap.Bool("local storage", isLocalStorage),
	)

	if !conf.Conf().Test.LocalUploadImage || isLocalStorage {
		sourceName := filepath.Base(filename)
		localPath := filepath.Join("cmd/web/static/images/brand_images", strings.ToLower(brand), sourceName)
		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			return "", err
		}
		data, err := io.ReadAll(file)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(localPath, data, 0644); err != nil {
			return "", err
		}
		return filename, nil
	}

	if err := s.objectStorage.PutObject(ctx, filename, file, contentType); err != nil {
		return "", err
	}

	key := s.objectStorage.GetPublicURL(filename)
	return key, nil
}

func (s *ImageService) Log() {
	logs.Log().Info("[ImageService] Loaded")
}

var _ IService = (*ImageService)(nil)

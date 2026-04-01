package services

import (
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

type ProductImageService struct {
	objectStorage storage.IObjectStorage
	encoder       encode.IEncode
	dbRO          database.IService
	dbRW          database.IService
}

func NewProductImageService(
	objectStorage storage.IObjectStorage,
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
) *ProductImageService {
	return &ProductImageService{
		objectStorage: objectStorage,
		encoder:       encoder,
		dbRO:          dbRO,
		dbRW:          dbRW,
	}
}

func (s *ProductImageService) GenerateFilename(
	ext string,
	paths ...string,
) string {
	uuid := utils.GenString(16)
	sanitizedName := strings.ReplaceAll(strings.Join(paths, "_"), " ", "_")
	return fmt.Sprintf("%s_%s%s", sanitizedName, uuid, ext)
}

func (s *ProductImageService) UploadProductImage(
	ctx context.Context,
	brand string,
	filename string,
	file io.Reader,
	contentType string,
) error {
	const logtag = "[ProductImageService]"
	isLocalStorage := s.objectStorage.ProviderEnum() == storage.STORAGE_PROVIDER_LOCAL

	logs.Log().Info(
		logtag,
		zap.String("uploading/storing file", filename),
		zap.Stringer("using", s.objectStorage.ProviderEnum()),
		zap.String("brand", brand),
	)

	if conf.Conf().Test.LocalUploadImage || isLocalStorage {
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

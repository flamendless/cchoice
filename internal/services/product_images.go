package services

import (
	"context"
	"fmt"
	"io"
	"strings"

	"cchoice/internal/database"
	"cchoice/internal/encode"
	"cchoice/internal/storage"
	"cchoice/internal/utils"
)

type ProductImageService struct {
	objectStorage storage.IObjectStorage
	encoder       encode.IEncode
	dbRO          database.Service
	dbRW          database.Service
}

func NewProductImageService(
	objectStorage storage.IObjectStorage,
	encoder encode.IEncode,
	dbRO database.Service,
	dbRW database.Service,
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
	return fmt.Sprintf("products/%s_%s%s", sanitizedName, uuid, ext)
}

func (s *ProductImageService) UploadProductImage(ctx context.Context, filename string, file io.Reader, contentType string) error {
	return s.objectStorage.PutObject(ctx, filename, file, contentType)
}

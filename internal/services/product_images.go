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

type ProductImagesService struct {
	objectStorage storage.IObjectStorage
	encoder       encode.IEncode
	dbRO          database.Service
	dbRW          database.Service
}

func NewProductImagesService(
	objectStorage storage.IObjectStorage,
	encoder encode.IEncode,
	dbRO database.Service,
	dbRW database.Service,
) *ProductImagesService {
	return &ProductImagesService{
		objectStorage: objectStorage,
		encoder:       encoder,
		dbRO:          dbRO,
		dbRW:          dbRW,
	}
}

func (s *ProductImagesService) GenerateFilename(
	ext string,
	paths ...string,
) string {
	uuid := utils.GenString(16)
	sanitizedName := strings.ReplaceAll(strings.Join(paths, "_"), " ", "_")
	return fmt.Sprintf("products/%s_%s%s", sanitizedName, uuid, ext)
}

func (s *ProductImagesService) UploadProductImage(ctx context.Context, filename string, file io.Reader, contentType string) error {
	return s.objectStorage.PutObject(ctx, filename, file, contentType)
}

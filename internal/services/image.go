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
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
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
	ip enums.ImagePrefix,
	ext string,
	paths ...string,
) string {
	if ip == enums.IMAGE_PREFIX_UNDEFINED {
		panic("RUNTIME. You must not pass undefined kind")
	}
	var dp string
	if !conf.Conf().IsProd() {
		if !strings.HasPrefix(dp, "DEV_") {
			dp = "DEV_"
		}
	}
	uuid := utils.GenString(16)
	sanitizedName := strings.ReplaceAll(strings.Join(paths, "_"), " ", "_")
	return fmt.Sprintf("%s%s_%s_%s%s", dp, ip.String(), sanitizedName, uuid, ext)
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
	if len(data) > int(constants.MaxSizeImageUpload) {
		return nil, fmt.Errorf("file too large: %d bytes", len(data))
	}
	return data, nil
}

func promoLocalImageURL(promoTitle, filename string) string {
	sourceName := filepath.Base(filename)
	promoDir := strings.ToLower(promoTitle)
	return utils.URL(fmt.Sprintf("/static/images/promo_images/%s/%s", promoDir, sourceName))
}

func (s *ImageService) savePromoImageLocally(promoTitle, filename string, data []byte) error {
	sourceName := filepath.Base(filename)
	localPath := filepath.Join("cmd/web/static/images/promo_images", strings.ToLower(promoTitle), sourceName)
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(localPath, data, 0644)
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

	if !conf.Conf().IsProd() && !conf.Conf().Test.LocalUploadImage || isLocalStorage {
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

	if !conf.Conf().IsProd() && !conf.Conf().Test.LocalUploadImage || isLocalStorage {
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
		return s.objectStorage.GetPublicURL(filename), nil
	}

	if err := s.objectStorage.PutObject(ctx, filename, file, contentType); err != nil {
		return "", err
	}

	key := s.objectStorage.GetPublicURL(filename)
	return key, nil
}

func (s *ImageService) UploadPromoBannerImage(
	ctx context.Context,
	promoTitle string,
	filename string,
	file io.Reader,
	contentType string,
) (string, error) {
	const logtag = "[ImageService] Promo Banner"
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
		zap.String("storing promo banner image", filename),
		zap.Stringer("using", s.objectStorage.ProviderEnum()),
		zap.String("promo title", promoTitle),
		zap.Bool("local storage", isLocalStorage),
	)

	if isLocalStorage {
		if err := s.savePromoImageLocally(promoTitle, filename, data); err != nil {
			return "", err
		}
		return promoLocalImageURL(promoTitle, filename), nil
	}

	if err := s.objectStorage.PutObject(ctx, filename, file, contentType); err != nil {
		return "", err
	}

	if !conf.Conf().IsProd() && !conf.Conf().Test.LocalUploadImage {
		if err := s.savePromoImageLocally(promoTitle, filename, data); err != nil {
			logs.Log().Warn(logtag, zap.Error(err))
		}
	}

	return s.objectStorage.GetPublicURL(filename), nil
}

// ValidateThemeLogoContentType restricts theme logo uploads to PNG and SVG
// only, since logos need transparency and SVG support (unlike product,
// brand, and promo images, which stay JPEG/PNG/WebP via ValidateContentType).
func (s *ImageService) ValidateThemeLogoContentType(contentType string) error {
	validTypes := map[string]bool{
		"image/png":     true,
		"image/svg+xml": true,
	}
	if !validTypes[contentType] {
		return fmt.Errorf("invalid content type: %s", contentType)
	}
	return nil
}

// BuildThemeLogoKey builds the Cloudflare object key for a theme logo, e.g.
// cchoice_local_theme_holidays_logowithtext.png
func BuildThemeLogoKey(themeTitle string, kind enums.ThemeLogoKind, ext string) string {
	env := strings.ToLower(conf.Conf().AppEnv.String())
	sanitizedTitle := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(themeTitle), " ", "_"))
	uuid := utils.GenString(16)
	return fmt.Sprintf("cchoice_%s_theme_%s_%s_%s%s", env, sanitizedTitle, kind.String(), uuid, strings.ToLower(ext))
}

func (s *ImageService) UploadThemeLogo(
	ctx context.Context,
	themeTitle string,
	kind enums.ThemeLogoKind,
	ext string,
	file io.Reader,
	contentType string,
) (string, error) {
	const logtag = "[ImageService] Theme Logo"
	if err := s.ValidateThemeLogoContentType(contentType); err != nil {
		return "", err
	}
	data, err := s.ValidateSize(file)
	if err != nil {
		return "", err
	}
	if int64(len(data)) > constants.MaxSizeThemeLogoUpload {
		return "", fmt.Errorf("file too large: %d bytes", len(data))
	}

	filename := BuildThemeLogoKey(themeTitle, kind, ext)
	file = bytes.NewReader(data)
	isLocalStorage := s.objectStorage.ProviderEnum() == storage.STORAGE_PROVIDER_LOCAL

	logs.Log().Info(
		logtag,
		zap.String("storing theme logo", filename),
		zap.Stringer("using", s.objectStorage.ProviderEnum()),
		zap.String("theme", themeTitle),
		zap.Stringer("kind", kind),
		zap.Bool("local storage", isLocalStorage),
	)

	if !conf.Conf().IsProd() && !conf.Conf().Test.LocalUploadImage || isLocalStorage {
		sourceName := filepath.Base(filename)
		localPath := filepath.Join("cmd/web/static/images/theme_logos", strings.ToLower(themeTitle), sourceName)
		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(localPath, data, 0644); err != nil {
			return "", err
		}
		return s.objectStorage.GetPublicURL(filename), nil
	}

	if err := s.objectStorage.PutObject(ctx, filename, file, contentType); err != nil {
		return "", err
	}

	return s.objectStorage.GetPublicURL(filename), nil
}

// ValidateInvoiceLogoContentType restricts invoice logo uploads to PNG and
// JPEG only, since these are the raster formats the PDF renderer can embed.
func (s *ImageService) ValidateInvoiceLogoContentType(contentType string) error {
	validTypes := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
	}
	if !validTypes[contentType] {
		return fmt.Errorf("invalid content type: %s", contentType)
	}
	return nil
}

// UploadInvoiceLogo stores the invoice/business logo and returns the public URL
// plus a local filesystem path (empty when stored on object storage) that the
// PDF renderer can read directly.
func (s *ImageService) UploadInvoiceLogo(
	ctx context.Context,
	ext string,
	file io.Reader,
	contentType string,
) (string, string, error) {
	const logtag = "[ImageService] Invoice Logo"
	if err := s.ValidateInvoiceLogoContentType(contentType); err != nil {
		return "", "", err
	}
	data, err := s.ValidateSize(file)
	if err != nil {
		return "", "", err
	}

	env := strings.ToLower(conf.Conf().AppEnv.String())
	uuid := utils.GenString(16)
	filename := fmt.Sprintf("cchoice_%s_invoice_logo_%s%s", env, uuid, strings.ToLower(ext))
	file = bytes.NewReader(data)
	isLocalStorage := s.objectStorage.ProviderEnum() == storage.STORAGE_PROVIDER_LOCAL

	logs.Log().Info(
		logtag,
		zap.String("storing invoice logo", filename),
		zap.Stringer("using", s.objectStorage.ProviderEnum()),
		zap.Bool("local storage", isLocalStorage),
	)

	if !conf.Conf().IsProd() && !conf.Conf().Test.LocalUploadImage || isLocalStorage {
		localPath := filepath.Join("cmd/web/static/images/invoice_logos", filepath.Base(filename))
		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			return "", "", err
		}
		if err := os.WriteFile(localPath, data, 0644); err != nil {
			return "", "", err
		}
		return s.objectStorage.GetPublicURL(filename), localPath, nil
	}

	if err := s.objectStorage.PutObject(ctx, filename, file, contentType); err != nil {
		return "", "", err
	}

	return s.objectStorage.GetPublicURL(filename), "", nil
}

func (s *ImageService) ID() string {
	return "Image"
}

func (s *ImageService) Log() {
	logs.Log().Info("[ImageService] Loaded")
}

var _ IService = (*ImageService)(nil)

package services

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"

	"cchoice/internal/constants"
	"cchoice/internal/encode/b64"
	"cchoice/internal/enums"
	"cchoice/internal/logs"

	"go.uber.org/zap"
)

type QRService struct {
	cache    *fastcache.Cache
	logoPath string
}

func NewQRService(cache *fastcache.Cache) *QRService {
	return &QRService{
		cache:    cache,
		logoPath: constants.PathLogoSmallLocal,
	}
}

func (s *QRService) GenerateQR(ctx context.Context, content string) ([]byte, error) {
	if s.cache != nil {
		if cached := s.cache.Get(nil, []byte(content)); len(cached) > 0 {
			return cached, nil
		}
	}

	qrc, err := qrcode.New(content)
	if err != nil {
		logs.LogCtx(ctx).Error("[QRService] failed to create qrcode", zap.Error(err), zap.String("content", content))
		return nil, fmt.Errorf("failed to create qrcode: %w", err)
	}

	options := []standard.ImageOption{
		standard.WithLogoImageFilePNG(s.logoPath),
	}

	buf := &bytes.Buffer{}
	w := standard.NewWithWriter(closerBuffer{buf}, options...)

	if err := qrc.Save(w); err != nil {
		logs.LogCtx(ctx).Error("[QRService] failed to save qrcode", zap.Error(err))
		return nil, fmt.Errorf("failed to save qrcode: %w", err)
	}

	if err := w.Close(); err != nil {
		logs.LogCtx(ctx).Error("[QRService] failed to close writer", zap.Error(err))
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	qrBytes := buf.Bytes()

	if s.cache != nil {
		s.cache.Set([]byte(content), qrBytes)
	}

	return qrBytes, nil
}

func (s *QRService) GenerateQRBase64(ctx context.Context, content string) (string, error) {
	qrBytes, err := s.GenerateQR(ctx, content)
	if err != nil {
		return "", err
	}
	imgfmt := enums.IMAGE_FORMAT_PNG.DataURIPrefix()
	return imgfmt + b64.ToBase64(qrBytes), nil
}

func (s *QRService) ID() string {
	return "QR"
}

func (s *QRService) Log() {
	logs.Log().Info("[QRService] Loaded")
}

var _ io.WriteCloser = closerBuffer{}
var _ IService = (*QRService)(nil)

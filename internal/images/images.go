package images

import (
	"cchoice/internal/encode/b64"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/metrics"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/VictoriaMetrics/fastcache"
	"go.uber.org/zap"
)

func ImageToB64(format ImageFormat, data []byte) string {
	return format.DataURIPrefix() + b64.ToBase64(data)
}

func GetImageDataB64(
	cache *fastcache.Cache,
	fs http.FileSystem,
	finalPath string,
	ext ImageFormat,
) (string, error) {
	const logtag = "[GetImageDataB64]"

	finalPath = strings.TrimPrefix(finalPath, "static")
	cacheKey := fmt.Appendf([]byte{}, "image_data_%s_%s", finalPath, ext.String())
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		metrics.Cache.MemHit()
		return string(data), nil
	} else {
		metrics.Cache.MemMiss()
	}

	f, err := fs.Open(finalPath)
	if err != nil {
		logs.Log().Error(logtag, zap.String("finalPath", finalPath), zap.Error(err))
		return "", errors.Join(errs.ErrFS, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logs.Log().Error("File close", zap.Error(err))
		}
	}()

	img, err := io.ReadAll(f)
	if err != nil {
		return "", errors.Join(errs.ErrFS, err)
	}

	imgData := ImageToB64(ext, img)
	cache.Set(cacheKey, []byte(imgData))
	return imgData, nil
}

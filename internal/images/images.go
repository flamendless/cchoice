package images

import (
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/serialize"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/VictoriaMetrics/fastcache"
	"go.uber.org/zap"
)

const (
	PNG  = "data:image/png;base64,"
	WEBP = "data:image/webp;base64,"
)

func ImageToB64(format enums.ImageFormat, data []byte) string {
	var base string
	switch format {
	case enums.IMAGE_FORMAT_PNG:
		base = PNG
	case enums.IMAGE_FORMAT_WEBP:
		base = WEBP
	default:
		panic("unhandled image format")
	}
	return base + serialize.ToBase64(data)
}

func GetImagePathWithSize(
	path string,
	size string,
	isThumbnail bool,
) (string, string, error) {
	ext := filepath.Ext(path)
	path = fmt.Sprintf("%s_%s%s", strings.TrimSuffix(path, ext), size, ext)
	if isThumbnail {
		path = strings.Replace(path, "/images/", "/thumbnails/", 1)
	}
	newPath, err := url.Parse(path)
	if err != nil {
		return "", "", err
	}
	return newPath.String(), ext, nil
}

func GetImageDataB64(
	cache *fastcache.Cache,
	fs http.FileSystem,
	finalPath string,
	ext string,
) (string, error) {
	cacheKey := fmt.Appendf([]byte{}, "image_data_%s_%s", finalPath, ext)
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		logs.Log().Debug("cache hit", zap.ByteString("key", cacheKey))
		return string(data), nil
	} else {
		logs.Log().Debug("cache miss", zap.ByteString("key", cacheKey))
	}

	f, err := fs.Open(finalPath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := f.Close(); err != nil {
			logs.Log().Error("File close", zap.Error(err))
		}
	}()

	img, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	imgData := ImageToB64(enums.ParseImageFormatExtToEnum(ext), img)
	cache.Set(cacheKey, []byte(imgData))
	return imgData, nil
}

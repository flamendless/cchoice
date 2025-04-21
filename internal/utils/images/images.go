package images

import (
	"cchoice/internal/constants"
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

func PNGEncode(data []byte) string {
	res := "data:image/png;base64,"
	res += serialize.ToBase64(data)
	return res
}

func WEBPEncode(data []byte) string {
	res := "data:image/webp;base64,"
	res += serialize.ToBase64(data)
	return res
}

func GetEncodedEmpty(fs http.FileSystem) (string, error) {
	f, err := fs.Open(constants.PathEmptyImage)
	if err != nil {
		return "", err
	}
	defer f.Close()

	img, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return WEBPEncode(img), nil
}

func GetThumbnailPath(path string, size string) (string, string, error) {
	ext := filepath.Ext(path)
	path = fmt.Sprintf("%s_%s%s", strings.TrimSuffix(path, ext), size, ext)
	path = strings.Replace(path, "/images/", "/thumbnails/", 1)
	newPath, err := url.Parse(path)
	if err != nil {
		return "", "", err
	}
	return newPath.String(), ext, nil
}

func GetImageData(
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
	defer f.Close()

	img, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	var imgData string
	switch ext {
	case ".webp":
		imgData = WEBPEncode(img)
	default:
		panic("unhandled ext")
	}

	cache.Set(cacheKey, []byte(imgData))
	return imgData, nil
}

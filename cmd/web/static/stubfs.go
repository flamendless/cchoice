//go:build !embeddedfs && !staticfs

package static

import (
	"io/fs"
	"net/http"
)

func GetMode() string {
	return "stubfs"
}

func GetFS() fs.FS {
	return nil
}

func Handler() http.Handler {
	return http.NotFoundHandler()
}

type fnCache func(http.ResponseWriter, *http.Request, http.FileSystem, string) (bool, http.File, error)

func CacheHandler(fn fnCache) http.Handler {
	return http.NotFoundHandler()
}

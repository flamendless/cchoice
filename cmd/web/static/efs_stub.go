//go:build !staticfs

package static

import (
	"io/fs"
	"net/http"
)

func GetFS() fs.FS {
	return nil
}

func Handler() http.Handler {
	return http.NotFoundHandler()
}

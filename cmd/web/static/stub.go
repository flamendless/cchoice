//go:build !embeddedfs && !staticfs

package static

import (
	"io/fs"
	"net/http"
)

func GetMode() string {
	return ""
}

func GetFS() fs.FS {
	return nil
}

func Handler() http.Handler {
	return http.NotFoundHandler()
}

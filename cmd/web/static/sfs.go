//go:build staticfs

package static

import (
	"io/fs"
	"net/http"
	"os"
)

const STATICFS = "./cmd/web/static"

func GetMode() string {
	return "staticfs"
}

func GetFS() fs.FS {
	return os.DirFS(STATICFS)
}

func Handler() http.Handler {
	root := http.Dir(STATICFS)
	return http.FileServer(root)
}

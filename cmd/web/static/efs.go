//go:build embeddedfs

package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed *
var files embed.FS

func GetMode() string {
	return "embeddedfs"
}

func GetFS() fs.FS {
	return files
}

func Handler() http.Handler {
	sub, err := fs.Sub(files, ".")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}

//go:build staticfs

package static

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
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

type fnCache func(http.ResponseWriter, *http.Request, http.FileSystem, string) (bool, http.File, error)

func CacheHandler(fn fnCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(STATICFS, r.URL.Path)

		notModified, file, err := fn(w, r, nil, path)
		if err != nil {
			fmt.Printf("[Static FS Cache Handler] %v\n", err)
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		if notModified {
			return
		}

		info, err := file.Stat()
		if err != nil {
			fmt.Printf("[Static FS Cache Handler] %v\n", err)
			http.NotFound(w, r)
			return
		}

		http.ServeContent(w, r, info.Name(), info.ModTime(), file)
	})
}

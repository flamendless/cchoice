package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed *
var Files embed.FS

func Handler() http.Handler {
	sub, err := fs.Sub(Files, ".")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}

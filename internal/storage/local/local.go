package local

import (
	"cchoice/cmd/web/static"
	"cchoice/internal/storage"
	"io/fs"
	"net/http"
	"strings"
)

type LocalFS struct {
	fs fs.FS
}

func New() storage.IFileSystem {
	fs := static.GetFS()
	if fs == nil {
		panic("static filesystem not initialized")
	}
	return &LocalFS{fs: fs}
}

func (l *LocalFS) Open(name string) (http.File, error) {
	name = strings.TrimPrefix(name, "/")
	name = strings.TrimPrefix(name, "static/")

	file, err := l.fs.Open(name)
	if err != nil {
		return nil, err
	}
	if httpFile, ok := file.(http.File); ok {
		return httpFile, nil
	}
	return http.FS(l.fs).Open(name)
}

var _ storage.IFileSystem = (*LocalFS)(nil)

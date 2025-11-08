package storage

import (
	"net/http"
)

type IFileSystem interface {
	http.FileSystem
}

type StorageProvider string

const (
	STORAGE_PROVIDER_LOCAL  StorageProvider = "local"
	STORAGE_PROVIDER_LINODE StorageProvider = "linode"
)

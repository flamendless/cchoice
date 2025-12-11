package storage

import (
	"cchoice/internal/errs"
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=StorageProvider -trimprefix=STORAGE_PROVIDER_

type StorageProvider int

const (
	STORAGE_PROVIDER_UNDEFINED StorageProvider = iota
	STORAGE_PROVIDER_LOCAL
	STORAGE_PROVIDER_LINODE
	STORAGE_PROVIDER_CLOUDFLARE_IMAGES
)

func ParseStorageProviderToEnum(sp string) StorageProvider {
	switch strings.ToUpper(sp) {
	case STORAGE_PROVIDER_LOCAL.String():
		return STORAGE_PROVIDER_LOCAL
	case STORAGE_PROVIDER_LINODE.String():
		return STORAGE_PROVIDER_LINODE
	case STORAGE_PROVIDER_CLOUDFLARE_IMAGES.String():
		return STORAGE_PROVIDER_CLOUDFLARE_IMAGES
	default:
		panic(fmt.Errorf("%w: '%s'", errs.ErrCmdUndefinedService, sp))
	}
}

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
)

func ParseStorageProviderToEnum(sp string) StorageProvider {
	switch strings.ToUpper(sp) {
	case STORAGE_PROVIDER_LOCAL.String():
		return STORAGE_PROVIDER_LOCAL
	case STORAGE_PROVIDER_LINODE.String():
		return STORAGE_PROVIDER_LINODE
	default:
		panic(fmt.Errorf("%w: '%s'", errs.ErrCmdUndefinedService, sp))
	}
}

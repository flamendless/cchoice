package storage

import (
	"context"
	"io"
	"net/http"
	"time"
)

type IFileSystem interface {
	http.FileSystem
}

type IObjectStorage interface {
	ProviderEnum() StorageProvider
	GetPublicURL(key string) string
	PresignedGetObject(ctx context.Context, key string, expiry time.Duration) (string, error)
	PutObject(ctx context.Context, key string, body io.Reader, contentType string) error
	PutObjectFromBytes(ctx context.Context, key string, data []byte, contentType string) error
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
	GetObjectBytes(ctx context.Context, key string) ([]byte, error)
	DeleteObject(ctx context.Context, key string) error
	ObjectExists(ctx context.Context, key string) (bool, error)
	HeadBucket(ctx context.Context) error
}

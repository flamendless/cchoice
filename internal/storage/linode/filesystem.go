package linode

import (
	"cchoice/internal/logs"
	"cchoice/internal/metrics"
	"cchoice/internal/storage"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
)

type LinodeFS struct {
	objstorage *Client
	bucket     string
	basePrefix string
}

type LinodeFile struct {
	body       io.ReadCloser
	name       string
	size       int64
	modTime    time.Time
	pos        int64
	isDir      bool
	objstorage *Client
	bucket     string
	key        string
}

type fileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func New(objstorage storage.IObjectStorage) storage.IFileSystem {
	linodeClient, ok := objstorage.(*Client)
	if !ok {
		panic("Invalid object storage")
	}

	ctx := context.Background()
	if err := linodeClient.HeadBucket(ctx); err != nil {
		panic(fmt.Errorf(
			"failed to connect to Linode bucket '%s': %w",
			linodeClient.GetBucket(),
			err,
		))
	}

	return &LinodeFS{
		objstorage: linodeClient,
		bucket:     linodeClient.GetBucket(),
		basePrefix: linodeClient.GetBasePrefix(),
	}
}

func (l *LinodeFS) Open(name string) (http.File, error) {
	const logtag = "[LinodeFS]"
	name = strings.TrimPrefix(name, "/")
	name = strings.TrimPrefix(name, "static/")

	key := name
	if l.basePrefix != "" {
		key = path.Join(l.basePrefix, name)
	}

	ctx := context.Background()

	body, err := l.objstorage.GetObject(ctx, name)
	if err != nil {
		metrics.Cache.LinodeAssetError()
		if strings.Contains(err.Error(), "NoSuchKey") {
			logs.Log().Debug(
				logtag,
				zap.String("error", "file not found"),
				zap.String("key", name),
				zap.String("bucket", l.bucket),
			)
			return nil, fmt.Errorf("file not found: %s", name)
		}
		logs.Log().Error(
			logtag,
			zap.Error(err),
			zap.String("key", name),
			zap.String("bucket", l.bucket),
		)
		return nil, fmt.Errorf("failed to get object from Linode: %w", err)
	}

	result, err := l.objstorage.HeadObject(ctx, name)
	var size int64
	modTime := time.Now()
	if err == nil {
		if result.ContentLength != nil {
			size = *result.ContentLength
		}
		if result.LastModified != nil {
			modTime = *result.LastModified
		}
	}

	metrics.Cache.LinodeAssetHit()
	logs.Log().Info(
		logtag,
		zap.String("action", "asset_retrieved"),
		zap.String("key", name),
		zap.String("bucket", l.bucket),
		zap.Int64("size", size),
	)

	return &LinodeFile{
		body:       body,
		name:       path.Base(name),
		size:       size,
		modTime:    modTime,
		isDir:      false,
		objstorage: l.objstorage,
		bucket:     l.bucket,
		key:        key,
	}, nil
}

func (f *LinodeFile) Read(p []byte) (n int, err error) {
	if f.body == nil {
		return 0, io.EOF
	}
	n, err = f.body.Read(p)
	f.pos += int64(n)
	return n, err
}

func (f *LinodeFile) Close() error {
	if f.body != nil {
		return f.body.Close()
	}
	return nil
}

func (f *LinodeFile) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("seek not supported for S3 objects")
}

func (f *LinodeFile) Readdir(count int) ([]fs.FileInfo, error) {
	if !f.isDir {
		return nil, errors.New("not a directory")
	}

	ctx := context.Background()
	prefix := f.key
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	objects, commonPrefixes, err := f.objstorage.ListObjectsV2(ctx, prefix, "/", int32(count))
	if err != nil {
		return nil, err
	}

	fileInfos := make([]fs.FileInfo, 0, len(objects)+len(commonPrefixes))
	for _, obj := range objects {
		info := &fileInfo{
			name:    path.Base(obj.Key),
			size:    obj.Size,
			modTime: *obj.LastModified,
			isDir:   false,
		}
		fileInfos = append(fileInfos, info)
	}

	for _, prefix := range commonPrefixes {
		key := prefix
		info := &fileInfo{
			name:    path.Base(strings.TrimSuffix(key, "/")),
			modTime: time.Now(),
			isDir:   true,
		}
		fileInfos = append(fileInfos, info)
	}

	return fileInfos, nil
}

func (f *LinodeFile) Stat() (fs.FileInfo, error) {
	return &fileInfo{
		name:    f.name,
		size:    f.size,
		modTime: f.modTime,
		isDir:   f.isDir,
	}, nil
}

func (f *fileInfo) Name() string       { return f.name }
func (f *fileInfo) Size() int64        { return f.size }
func (f *fileInfo) Mode() os.FileMode  { return 0644 }
func (f *fileInfo) ModTime() time.Time { return f.modTime }
func (f *fileInfo) IsDir() bool        { return f.isDir }
func (f *fileInfo) Sys() interface{}   { return nil }

var _ storage.IFileSystem = (*LinodeFS)(nil)

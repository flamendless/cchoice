package linode

import (
	"cchoice/internal/storage"
	s3client "cchoice/internal/storage/s3"
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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type LinodeFS struct {
	s3Client   *s3client.Client
	bucket     string
	basePrefix string
}

type LinodeFile struct {
	body     io.ReadCloser
	name     string
	size     int64
	modTime  time.Time
	pos      int64
	isDir    bool
	s3Client *s3client.Client
	bucket   string
	key      string
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

	listInput := &s3.ListObjectsV2Input{
		Bucket:    aws.String(f.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int32(int32(count)),
	}

	result, err := f.s3Client.GetS3Client().ListObjectsV2(ctx, listInput)
	if err != nil {
		return nil, err
	}

	fileInfos := make([]fs.FileInfo, 0, len(result.Contents))
	for _, obj := range result.Contents {
		key := *obj.Key
		info := &fileInfo{
			name:    path.Base(key),
			size:    *obj.Size,
			modTime: *obj.LastModified,
			isDir:   false,
		}
		fileInfos = append(fileInfos, info)
	}

	for _, prefix := range result.CommonPrefixes {
		key := *prefix.Prefix
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

type fileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (f *fileInfo) Name() string       { return f.name }
func (f *fileInfo) Size() int64        { return f.size }
func (f *fileInfo) Mode() os.FileMode  { return 0644 }
func (f *fileInfo) ModTime() time.Time { return f.modTime }
func (f *fileInfo) IsDir() bool        { return f.isDir }
func (f *fileInfo) Sys() interface{}   { return nil }

func New() storage.IFileSystem {
	s3Client, err := s3client.NewClientFromConfig()
	if err != nil {
		panic(fmt.Errorf("failed to create S3 client: %w", err))
	}

	ctx := context.Background()
	if err := s3Client.HeadBucket(ctx); err != nil {
		panic(fmt.Errorf("failed to connect to Linode bucket '%s': %w", s3Client.GetBucket(), err))
	}

	return &LinodeFS{
		s3Client:   s3Client,
		bucket:     s3Client.GetBucket(),
		basePrefix: s3Client.GetBasePrefix(),
	}
}

func (l *LinodeFS) Open(name string) (http.File, error) {
	name = strings.TrimPrefix(name, "/")
	name = strings.TrimPrefix(name, "static/")

	key := name
	if l.basePrefix != "" {
		key = path.Join(l.basePrefix, name)
	}

	ctx := context.Background()

	body, err := l.s3Client.GetObject(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, fmt.Errorf("file not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get object from Linode: %w", err)
	}

	result, err := l.s3Client.HeadObject(ctx, name)
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

	return &LinodeFile{
		body:     body,
		name:     path.Base(name),
		size:     size,
		modTime:  modTime,
		isDir:    false,
		s3Client: l.s3Client,
		bucket:   l.bucket,
		key:      key,
	}, nil
}

var _ storage.IFileSystem = (*LinodeFS)(nil)

package linode

import (
	"cchoice/internal/conf"
	"cchoice/internal/errs"
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

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type LinodeFS struct {
	client     *s3.Client
	bucket     string
	basePrefix string
}

type LinodeFile struct {
	body    io.ReadCloser
	name    string
	size    int64
	modTime time.Time
	pos     int64
	isDir   bool
	client  *s3.Client
	bucket  string
	key     string
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

	result, err := f.client.ListObjectsV2(ctx, listInput)
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
	cfg := conf.Conf()
	if cfg.StorageProvider != "linode" {
		panic("STORAGE_PROVIDER must be 'linode' to use LinodeFS")
	}

	if cfg.Linode.Endpoint == "" || cfg.Linode.AccessKey == "" || cfg.Linode.SecretKey == "" || cfg.Linode.Bucket == "" {
		panic(fmt.Errorf("[Linode]: %w", errs.ErrEnvVarRequired))
	}

	ctx := context.Background()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Linode.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.Linode.AccessKey,
			cfg.Linode.SecretKey,
			"",
		)),
	)
	if err != nil {
		panic(fmt.Errorf("failed to load AWS config: %w", err))
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Linode.Endpoint)
		o.UsePathStyle = true
	})

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.Linode.Bucket),
	})
	if err != nil {
		panic(fmt.Errorf("failed to connect to Linode bucket '%s': %w", cfg.Linode.Bucket, err))
	}

	return &LinodeFS{
		client:     client,
		bucket:     cfg.Linode.Bucket,
		basePrefix: strings.TrimPrefix(cfg.Linode.BasePrefix, "/"),
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

	getInput := &s3.GetObjectInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(key),
	}

	result, err := l.client.GetObject(ctx, getInput)
	if err != nil {
		var apiErr interface{ ErrorCode() string }
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchKey" {
			return nil, fmt.Errorf("file not found: %s", name)
		}
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, fmt.Errorf("file not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get object from Linode: %w", err)
	}

	var size int64
	if result.ContentLength != nil {
		size = *result.ContentLength
	}

	modTime := time.Now()
	if result.LastModified != nil {
		modTime = *result.LastModified
	}

	return &LinodeFile{
		body:    result.Body,
		name:    path.Base(name),
		size:    size,
		modTime: modTime,
		isDir:   false,
		client:  l.client,
		bucket:  l.bucket,
		key:     key,
	}, nil
}

var _ storage.IFileSystem = (*LinodeFS)(nil)

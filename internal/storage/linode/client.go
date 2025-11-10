package linode

import (
	"bytes"
	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/storage"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type Client struct {
	minioClient *minio.Client
	bucket      string
	basePrefix  string
	endpoint    string
}

type Config struct {
	Endpoint   string
	Region     string
	AccessKey  string
	SecretKey  string
	Bucket     string
	BasePrefix string
	APIToken   string
}

type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified *time.Time
}

type HeadObjectOutput struct {
	ContentLength *int64
	LastModified  *time.Time
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.Endpoint == "" || cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Bucket == "" {
		return nil, errs.ErrEnvVarRequired
	}

	endpoint := cfg.Endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	endpoint = strings.TrimSuffix(endpoint, "/")

	host := strings.TrimPrefix(endpoint, "https://")
	host = strings.TrimPrefix(host, "http://")

	useSSL := strings.HasPrefix(endpoint, "https://")

	minioClient, err := minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: useSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	basePrefix := strings.TrimPrefix(cfg.BasePrefix, "/")

	return &Client{
		minioClient: minioClient,
		bucket:      cfg.Bucket,
		basePrefix:  basePrefix,
		endpoint:    endpoint,
	}, nil
}

func NewClientFromConfig() (*Client, error) {
	cfg := conf.Conf()
	if cfg.StorageProvider != "linode" {
		return nil, errs.ErrLinodeServiceInit
	}

	logs.Log().Info(
		"Linode keys",
		zap.String("endpoint", cfg.Linode.Endpoint),
		zap.String("bucket", cfg.Linode.Bucket),
		zap.String("region", cfg.Linode.Region),
		zap.String("base prefix", cfg.Linode.BasePrefix),
	)

	return NewClient(Config{
		Endpoint:   cfg.Linode.Endpoint,
		Region:     cfg.Linode.Region,
		AccessKey:  cfg.Linode.AccessKey,
		SecretKey:  cfg.Linode.SecretKey,
		Bucket:     cfg.Linode.Bucket,
		BasePrefix: cfg.Linode.BasePrefix,
	})
}

func MustInit() storage.IObjectStorage {
	cfg := conf.Conf()
	if cfg.StorageProvider != "linode" {
		panic("'STORAGE_PROVIDER' must be 'linode' to use this")
	}

	client, err := NewClientFromConfig()
	if err != nil {
		panic(errors.Join(errs.ErrLinodeServiceInit, fmt.Errorf("failed to initialize Linode client: %w", err)))
	}

	ctx := context.Background()
	if err := client.HeadBucket(ctx); err != nil {
		panic(errors.Join(errs.ErrLinodeServiceInit, fmt.Errorf("failed to connect to Linode bucket: %w", err)))
	}

	return client
}

func (c *Client) normalizeKey(key string) string {
	key = strings.TrimPrefix(key, "/")
	key = strings.TrimPrefix(key, "static/")

	if c.basePrefix != "" {
		key = c.basePrefix + "/" + key
		key = strings.ReplaceAll(key, "//", "/")
	}

	return key
}

func (c *Client) HeadBucket(ctx context.Context) error {
	exists, err := c.minioClient.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("bucket '%s' does not exist", c.bucket)
	}
	return nil
}

func (c *Client) PutObject(ctx context.Context, key string, body io.Reader, contentType string) error {
	normalizedKey := c.normalizeKey(key)

	_, err := c.minioClient.PutObject(ctx, c.bucket, normalizedKey, body, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to put object '%s': %w", normalizedKey, err)
	}

	return nil
}

func (c *Client) PutObjectFromBytes(ctx context.Context, key string, data []byte, contentType string) error {
	return c.PutObject(ctx, key, bytes.NewReader(data), contentType)
}

func (c *Client) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	normalizedKey := c.normalizeKey(key)

	obj, err := c.minioClient.GetObject(ctx, c.bucket, normalizedKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object '%s': %w", normalizedKey, err)
	}

	return obj, nil
}

func (c *Client) GetObjectBytes(ctx context.Context, key string) ([]byte, error) {
	body, err := c.GetObject(ctx, key)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return io.ReadAll(body)
}

func (c *Client) DeleteObject(ctx context.Context, key string) error {
	normalizedKey := c.normalizeKey(key)

	err := c.minioClient.RemoveObject(ctx, c.bucket, normalizedKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object '%s': %w", normalizedKey, err)
	}

	return nil
}

func (c *Client) ListObjects(ctx context.Context, prefix string, maxKeys int32) ([]ObjectInfo, error) {
	normalizedPrefix := c.normalizeKey(prefix)

	objectCh := c.minioClient.ListObjects(ctx, c.bucket, minio.ListObjectsOptions{
		Prefix:    normalizedPrefix,
		Recursive: false,
		MaxKeys:   int(maxKeys),
	})

	objects := make([]ObjectInfo, 0, len(objectCh))
	for obj := range objectCh {
		if obj.Err != nil {
			return nil, fmt.Errorf("failed to list objects with prefix '%s': %w", normalizedPrefix, obj.Err)
		}
		objects = append(objects, ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: &obj.LastModified,
		})
	}

	return objects, nil
}

func (c *Client) ListObjectsV2(ctx context.Context, prefix string, delimiter string, maxKeys int32) ([]ObjectInfo, []string, error) {
	normalizedPrefix := c.normalizeKey(prefix)

	objectCh := c.minioClient.ListObjects(ctx, c.bucket, minio.ListObjectsOptions{
		Prefix:    normalizedPrefix,
		Recursive: delimiter == "",
		MaxKeys:   int(maxKeys),
	})

	var objects []ObjectInfo
	var commonPrefixes []string
	prefixMap := make(map[string]bool)

	for obj := range objectCh {
		if obj.Err != nil {
			return nil, nil, fmt.Errorf("failed to list objects with prefix '%s': %w", normalizedPrefix, obj.Err)
		}

		if delimiter != "" && strings.Contains(obj.Key[len(normalizedPrefix):], delimiter) {
			parts := strings.Split(strings.TrimPrefix(obj.Key, normalizedPrefix), delimiter)
			if len(parts) > 0 {
				prefix := normalizedPrefix + parts[0] + delimiter
				if !prefixMap[prefix] {
					prefixMap[prefix] = true
					commonPrefixes = append(commonPrefixes, prefix)
				}
			}
		} else {
			objects = append(objects, ObjectInfo{
				Key:          obj.Key,
				Size:         obj.Size,
				LastModified: &obj.LastModified,
			})
		}
	}

	return objects, commonPrefixes, nil
}

func (c *Client) HeadObject(ctx context.Context, key string) (*HeadObjectOutput, error) {
	normalizedKey := c.normalizeKey(key)

	objInfo, err := c.minioClient.StatObject(ctx, c.bucket, normalizedKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to head object '%s': %w", normalizedKey, err)
	}

	return &HeadObjectOutput{
		ContentLength: &objInfo.Size,
		LastModified:  &objInfo.LastModified,
	}, nil
}

func (c *Client) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := c.HeadObject(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (c *Client) GetBucket() string {
	return c.bucket
}

func (c *Client) GetBasePrefix() string {
	return c.basePrefix
}

func (c *Client) GetMinioClient() *minio.Client {
	return c.minioClient
}

func (c *Client) ProviderEnum() storage.StorageProvider {
	return storage.STORAGE_PROVIDER_LINODE
}

func (c *Client) GetPublicURL(key string) string {
	normalizedKey := c.normalizeKey(key)
	endpoint := strings.TrimPrefix(c.endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	normalizedKey = strings.TrimPrefix(normalizedKey, "/")
	url := fmt.Sprintf("https://%s.%s/%s", c.bucket, endpoint, normalizedKey)
	return url
}

func (c *Client) PresignedGetObject(ctx context.Context, key string, expiry time.Duration) (string, error) {
	normalizedKey := c.normalizeKey(key)
	url, err := c.minioClient.PresignedGetObject(ctx, c.bucket, normalizedKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL for '%s': %w", normalizedKey, err)
	}
	return url.String(), nil
}

var _ storage.IObjectStorage = (*Client)(nil)

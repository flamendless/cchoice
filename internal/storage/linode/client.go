package linode

import (
	"bytes"
	"cchoice/internal/conf"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/storage"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type Client struct {
	minioClient    *minio.Client
	bucketEnum     enums.LinodeBucketEnum
	basePrefix     string
	endpoint       string
	urlCache       sync.Map
	presignedCache sync.Map
}

type presignedCacheEntry struct {
	url    string
	expiry time.Time
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

func validate() {
	cfg := conf.Conf()
	if cfg.Linode.Endpoint == "" || cfg.Linode.Region == "" {
		panic(fmt.Errorf("[Linode Storage]: %w", errs.ErrEnvVarRequired))
	}

	buckets := cfg.Linode.GetBuckets()
	for bucketEnum, bucketConfig := range buckets {
		if bucketConfig.Bucket == "" {
			continue
		}
		if bucketConfig.AccessKey == "" {
			panic(fmt.Errorf("[Linode Storage %s]: access key must be configure", bucketEnum.String()))
		}
		if bucketConfig.SecretKey == "" {
			panic(fmt.Errorf("[Linode Storage %s]: secret key must be configured", bucketEnum.String()))
		}
	}

	if len(buckets) != 2 {
		panic("[Linode Storage]: exactly two buckets must be configured")
	}
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
		bucketEnum:  enums.LINODE_BUCKET_UNDEFINED,
		basePrefix:  basePrefix,
		endpoint:    endpoint,
	}, nil
}

func NewClientFromConfig() (*Client, error) {
	return NewClientFromConfigWithBucket(enums.LINODE_BUCKET_PRIVATE)
}

func NewClientFromConfigWithBucket(bucketEnum enums.LinodeBucketEnum) (*Client, error) {
	cfg := conf.Conf()
	if cfg.StorageProvider != storage.STORAGE_PROVIDER_LINODE.String() {
		return nil, errs.ErrLinodeServiceInit
	}

	bucketConfig, ok := cfg.Linode.GetBucketConfig(bucketEnum)
	if !ok {
		return nil, fmt.Errorf("bucket config for enum %s not found", bucketEnum.String())
	}

	if bucketConfig.Bucket == "" {
		return nil, fmt.Errorf("bucket for enum %s is not configured", bucketEnum.String())
	}
	if bucketConfig.AccessKey == "" {
		return nil, fmt.Errorf("access key for bucket enum %s is not configured", bucketEnum.String())
	}
	if bucketConfig.SecretKey == "" {
		return nil, fmt.Errorf("secret key for bucket enum %s is not configured", bucketEnum.String())
	}

	logs.Log().Info(
		"New client",
		zap.String("endpoint", cfg.Linode.Endpoint),
		zap.String("bucket", bucketConfig.Bucket),
		zap.String("region", cfg.Linode.Region),
		zap.String("base prefix", cfg.Linode.BasePrefix),
	)

	client, err := NewClient(Config{
		Endpoint:   cfg.Linode.Endpoint,
		Region:     cfg.Linode.Region,
		AccessKey:  bucketConfig.AccessKey,
		SecretKey:  bucketConfig.SecretKey,
		Bucket:     bucketConfig.Bucket,
		BasePrefix: cfg.Linode.BasePrefix,
	})
	if err != nil {
		return nil, err
	}
	client.bucketEnum = bucketEnum
	return client, nil
}

func MustInit() storage.IObjectStorage {
	return MustInitWithBucket(enums.LINODE_BUCKET_PRIVATE)
}

func MustInitWithBucket(bucketEnum enums.LinodeBucketEnum) storage.IObjectStorage {
	validate()
	cfg := conf.Conf()
	if cfg.StorageProvider != storage.STORAGE_PROVIDER_LINODE.String() {
		panic("'STORAGE_PROVIDER' must be 'linode' to use this")
	}

	client, err := NewClientFromConfigWithBucket(bucketEnum)
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
	bucket := c.GetBucket()
	exists, err := c.minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("bucket '%s' does not exist", bucket)
	}
	return nil
}

func (c *Client) PutObject(ctx context.Context, key string, body io.Reader, contentType string) error {
	normalizedKey := c.normalizeKey(key)
	bucket := c.GetBucket()

	_, err := c.minioClient.PutObject(ctx, bucket, normalizedKey, body, -1, minio.PutObjectOptions{
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
	bucket := c.GetBucket()

	obj, err := c.minioClient.GetObject(ctx, bucket, normalizedKey, minio.GetObjectOptions{})
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
	bucket := c.GetBucket()

	err := c.minioClient.RemoveObject(ctx, bucket, normalizedKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object '%s': %w", normalizedKey, err)
	}

	return nil
}

func (c *Client) ListObjects(ctx context.Context, prefix string, maxKeys int32) ([]ObjectInfo, error) {
	normalizedPrefix := c.normalizeKey(prefix)
	bucket := c.GetBucket()

	objectCh := c.minioClient.ListObjects(ctx, bucket, minio.ListObjectsOptions{
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
	bucket := c.GetBucket()

	objectCh := c.minioClient.ListObjects(ctx, bucket, minio.ListObjectsOptions{
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
	bucket := c.GetBucket()

	objInfo, err := c.minioClient.StatObject(ctx, bucket, normalizedKey, minio.StatObjectOptions{})
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
	cfg := conf.Conf()
	bucketConfig, ok := cfg.Linode.GetBucketConfig(c.bucketEnum)
	if !ok {
		return ""
	}
	return bucketConfig.Bucket
}

func (c *Client) GetBucketEnum() enums.LinodeBucketEnum {
	return c.bucketEnum
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
	if cachedURL, ok := c.urlCache.Load(key); ok {
		return cachedURL.(string)
	}

	normalizedKey := c.normalizeKey(key)
	bucket := c.GetBucket()
	endpoint := strings.TrimPrefix(c.endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	normalizedKey = strings.TrimPrefix(normalizedKey, "/")
	url := fmt.Sprintf("https://%s.%s/%s", bucket, endpoint, normalizedKey)

	c.urlCache.Store(key, url)
	return url
}

func (c *Client) PresignedGetObject(ctx context.Context, key string, expiry time.Duration) (string, error) {
	cacheKey := fmt.Sprintf("%s:%d", key, int64(expiry.Seconds()))
	if cached, ok := c.presignedCache.Load(cacheKey); ok {
		entry := cached.(presignedCacheEntry)
		if time.Now().Before(entry.expiry) {
			return entry.url, nil
		}
		c.presignedCache.Delete(cacheKey)
	}

	normalizedKey := c.normalizeKey(key)
	bucket := c.GetBucket()
	url, err := c.minioClient.PresignedGetObject(ctx, bucket, normalizedKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL for '%s': %w", normalizedKey, err)
	}

	urlStr := url.String()
	cacheEntry := presignedCacheEntry{
		url:    urlStr,
		expiry: time.Now().Add(expiry - time.Minute),
	}
	c.presignedCache.Store(cacheKey, cacheEntry)

	return urlStr, nil
}

var _ storage.IObjectStorage = (*Client)(nil)

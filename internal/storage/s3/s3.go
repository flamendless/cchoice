package s3

import (
	"bytes"
	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"
)

type Client struct {
	s3Client   *s3.Client
	bucket     string
	basePrefix string
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

func NewClient(cfg Config) (*Client, error) {
	if cfg.Endpoint == "" || cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Bucket == "" {
		return nil, errs.ErrEnvVarRequired
	}

	ctx := context.Background()

	endpoint := cfg.Endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	endpoint = strings.TrimSuffix(endpoint, "/")
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	basePrefix := strings.TrimPrefix(cfg.BasePrefix, "/")

	return &Client{
		s3Client:   s3Client,
		bucket:     cfg.Bucket,
		basePrefix: basePrefix,
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
	_, err := c.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.bucket),
	})
	return err
}

func (c *Client) PutObject(ctx context.Context, key string, body io.Reader, contentType string) error {
	normalizedKey := c.normalizeKey(key)

	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(normalizedKey),
		Body:        body,
		ContentType: aws.String(contentType),
	}

	_, err := c.s3Client.PutObject(ctx, putInput)
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

	getInput := &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(normalizedKey),
	}

	result, err := c.s3Client.GetObject(ctx, getInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get object '%s': %w", normalizedKey, err)
	}

	return result.Body, nil
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

	deleteInput := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(normalizedKey),
	}

	_, err := c.s3Client.DeleteObject(ctx, deleteInput)
	if err != nil {
		return fmt.Errorf("failed to delete object '%s': %w", normalizedKey, err)
	}

	return nil
}

func (c *Client) ListObjects(ctx context.Context, prefix string, maxKeys int32) ([]types.Object, error) {
	normalizedPrefix := c.normalizeKey(prefix)

	listInput := &s3.ListObjectsV2Input{
		Bucket:  aws.String(c.bucket),
		Prefix:  aws.String(normalizedPrefix),
		MaxKeys: aws.Int32(maxKeys),
	}

	result, err := c.s3Client.ListObjectsV2(ctx, listInput)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects with prefix '%s': %w", normalizedPrefix, err)
	}

	return result.Contents, nil
}

func (c *Client) HeadObject(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	normalizedKey := c.normalizeKey(key)

	headInput := &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(normalizedKey),
	}

	result, err := c.s3Client.HeadObject(ctx, headInput)
	if err != nil {
		return nil, fmt.Errorf("failed to head object '%s': %w", normalizedKey, err)
	}

	return result, nil
}

func (c *Client) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := c.HeadObject(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
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

func (c *Client) GetS3Client() *s3.Client {
	return c.s3Client
}

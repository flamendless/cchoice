package cloudflare

import (
	"bytes"
	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	cloudflareAPIBaseURL = "https://api.cloudflare.com/client/v4"
	imageDeliveryBaseURL = "https://imagedelivery.net"
)

type Client struct {
	httpClient  *http.Client
	accountID   string
	accountHash string
	apiToken    string
	variant     string
	urlCache    sync.Map
}

type Config struct {
	AccountID   string
	AccountHash string
	APIToken    string
	Variant     string
}

type CloudflareResponse struct {
	Success  bool                   `json:"success"`
	Errors   []CloudflareError      `json:"errors"`
	Messages []string               `json:"messages"`
	Result   map[string]interface{} `json:"result"`
}

type CloudflareError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ImageUploadResult struct {
	ID       string   `json:"id"`
	Filename string   `json:"filename"`
	Uploaded string   `json:"uploaded"`
	Variants []string `json:"variants"`
}

type ImageDetailsResult struct {
	ID       string   `json:"id"`
	Filename string   `json:"filename"`
	Uploaded string   `json:"uploaded"`
	Variants []string `json:"variants"`
}

func validate() {
	cfg := conf.Conf()
	if cfg.StorageProvider != storage.STORAGE_PROVIDER_CLOUDFLARE_IMAGES.String() {
		return
	}

	if cfg.CloudflareImages.AccountID == "" {
		panic(errs.ErrCloudflareAccountID)
	}
	if cfg.CloudflareImages.AccountHash == "" {
		panic(errs.ErrCloudflareAccountHash)
	}
	if cfg.CloudflareImages.APIToken == "" {
		panic(errs.ErrCloudflareAPIToken)
	}
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.AccountID == "" || cfg.AccountHash == "" || cfg.APIToken == "" {
		return nil, errs.ErrCloudflareServiceInit
	}

	variant := cfg.Variant
	if variant == "" {
		variant = "public"
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		accountID:   cfg.AccountID,
		accountHash: cfg.AccountHash,
		apiToken:    cfg.APIToken,
		variant:     variant,
	}, nil
}

func NewClientFromConfig() (*Client, error) {
	cfg := conf.Conf()
	if cfg.StorageProvider != storage.STORAGE_PROVIDER_CLOUDFLARE_IMAGES.String() {
		return nil, errs.ErrCloudflareServiceInit
	}

	return NewClient(Config{
		AccountID:   cfg.CloudflareImages.AccountID,
		AccountHash: cfg.CloudflareImages.AccountHash,
		APIToken:    cfg.CloudflareImages.APIToken,
		Variant:     cfg.CloudflareImages.Variant,
	})
}

func MustInit() storage.IObjectStorage {
	validate()
	cfg := conf.Conf()
	if cfg.StorageProvider != storage.STORAGE_PROVIDER_CLOUDFLARE_IMAGES.String() {
		panic(errs.ErrCloudflareServiceInit)
	}

	client, err := NewClientFromConfig()
	if err != nil {
		panic(errors.Join(errs.ErrCloudflareServiceInit, err))
	}

	logs.Log().Info(
		"Cloudflare Images",
		zap.String("account_id", cfg.CloudflareImages.AccountID),
		zap.String("variant", cfg.CloudflareImages.Variant),
	)

	ctx := context.Background()
	if err := client.HeadBucket(ctx); err != nil {
		panic(errors.Join(errs.ErrCloudflareVerifyAccess, err))
	}

	return client
}

// normalizeKey converts a file path to a Cloudflare Image ID
// e.g., "static/images/brand_logos/BOSCH.webp" -> "brand_logos-BOSCH"
func (c *Client) normalizeKey(key string) string {
	key = strings.TrimPrefix(key, "/")
	key = strings.TrimPrefix(key, "static/")
	key = strings.TrimPrefix(key, "images/")
	ext := filepath.Ext(key)
	key = strings.TrimSuffix(key, ext)
	key = strings.ReplaceAll(key, "/", "-")
	key = strings.ReplaceAll(key, "\\", "-")
	return key
}

func (c *Client) ProviderEnum() storage.StorageProvider {
	return storage.STORAGE_PROVIDER_CLOUDFLARE_IMAGES
}

func (c *Client) GetPublicURL(key string) string {
	if cachedURL, ok := c.urlCache.Load(key); ok {
		return cachedURL.(string)
	}

	imageID := c.normalizeKey(key)
	url := fmt.Sprintf("%s/%s/%s/%s", imageDeliveryBaseURL, c.accountHash, imageID, c.variant)

	c.urlCache.Store(key, url)
	return url
}

func (c *Client) PresignedGetObject(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return c.GetPublicURL(key), nil
}

func (c *Client) PutObject(ctx context.Context, key string, body io.Reader, contentType string) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}
	return c.PutObjectFromBytes(ctx, key, data, contentType)
}

func (c *Client) PutObjectFromBytes(ctx context.Context, key string, data []byte, contentType string) error {
	imageID := c.normalizeKey(key)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if err := writer.WriteField("id", imageID); err != nil {
		return fmt.Errorf("failed to write id field: %w", err)
	}

	filename := filepath.Base(key)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(data); err != nil {
		return fmt.Errorf("failed to write file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	url := fmt.Sprintf("%s/accounts/%s/images/v1", cloudflareAPIBaseURL, c.accountID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Join(errs.ErrCloudflareUpload, err)
	}
	defer resp.Body.Close()

	var cfResp CloudflareResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return errors.Join(errs.ErrCloudflareUpload, err)
	}

	if !cfResp.Success {
		errMsgs := make([]string, len(cfResp.Errors))
		for i, e := range cfResp.Errors {
			errMsgs[i] = e.Message
		}
		return errors.Join(errs.ErrCloudflareAPI, fmt.Errorf("%s", strings.Join(errMsgs, "; ")))
	}

	return nil
}

func (c *Client) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	url := c.GetPublicURL(key)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Join(errs.ErrCloudflareGet, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Join(errs.ErrCloudflareGet, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.Join(errs.ErrCloudflareGet, fmt.Errorf("status %d", resp.StatusCode))
	}

	return resp.Body, nil
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
	imageID := c.normalizeKey(key)
	url := fmt.Sprintf("%s/accounts/%s/images/v1/%s", cloudflareAPIBaseURL, c.accountID, imageID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return errors.Join(errs.ErrCloudflareDelete, err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Join(errs.ErrCloudflareDelete, err)
	}
	defer resp.Body.Close()

	var cfResp CloudflareResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return errors.Join(errs.ErrCloudflareDelete, err)
	}

	if !cfResp.Success {
		errMsgs := make([]string, len(cfResp.Errors))
		for i, e := range cfResp.Errors {
			errMsgs[i] = e.Message
		}
		return errors.Join(errs.ErrCloudflareAPI, fmt.Errorf("%s", strings.Join(errMsgs, "; ")))
	}

	return nil
}

func (c *Client) ObjectExists(ctx context.Context, key string) (bool, error) {
	imageID := c.normalizeKey(key)
	url := fmt.Sprintf("%s/accounts/%s/images/v1/%s", cloudflareAPIBaseURL, c.accountID, imageID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, errors.Join(errs.ErrCloudflareAPI, err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, errors.Join(errs.ErrCloudflareAPI, err)
	}
	defer resp.Body.Close()

	var cfResp CloudflareResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return false, errors.Join(errs.ErrCloudflareAPI, err)
	}

	return cfResp.Success, nil
}

func (c *Client) HeadBucket(ctx context.Context) error {
	url := fmt.Sprintf("%s/accounts/%s/images/v1?per_page=1", cloudflareAPIBaseURL, c.accountID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.Join(errs.ErrCloudflareVerifyAccess, err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Join(errs.ErrCloudflareVerifyAccess, err)
	}
	defer resp.Body.Close()

	var cfResp CloudflareResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return errors.Join(errs.ErrCloudflareVerifyAccess, err)
	}

	if !cfResp.Success {
		errMsgs := make([]string, len(cfResp.Errors))
		for i, e := range cfResp.Errors {
			errMsgs[i] = e.Message
		}
		return errors.Join(errs.ErrCloudflareAPI, fmt.Errorf("%s", strings.Join(errMsgs, "; ")))
	}

	return nil
}

type BatchUploadImage struct {
	Key         string
	Data        []byte
	ContentType string
}

func (c *Client) BatchUpload(ctx context.Context, images []BatchUploadImage) ([]string, []error) {
	uploadedIDs := make([]string, 0, len(images))
	uploadErrs := make([]error, 0)

	for _, img := range images {
		if err := c.PutObjectFromBytes(ctx, img.Key, img.Data, img.ContentType); err != nil {
			uploadErrs = append(uploadErrs, errors.Join(errs.ErrCloudflareUpload, err))
		} else {
			uploadedIDs = append(uploadedIDs, c.normalizeKey(img.Key))
		}
	}

	return uploadedIDs, uploadErrs
}

func (c *Client) GetAccountID() string {
	return c.accountID
}

func (c *Client) GetAccountHash() string {
	return c.accountHash
}

func (c *Client) GetVariant() string {
	return c.variant
}

var _ storage.IObjectStorage = (*Client)(nil)

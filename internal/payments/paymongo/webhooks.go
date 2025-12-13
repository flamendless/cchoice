package paymongo

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"cchoice/internal/logs"

	"go.uber.org/zap"
)

const (
	WebhookEventPaymentPaid                = "payment.paid"
	WebhookEventPaymentFailed              = "payment.failed"
	WebhookEventPaymentRefunded            = "payment.refunded"
	WebhookEventPaymentRefundUpdated       = "payment.refund.updated"
	WebhookEventCheckoutSessionPaymentPaid = "checkout_session.payment.paid"
	WebhookEventLinkPaymentPaid            = "link.payment.paid"
	WebhookEventSourceChargeable           = "source.chargeable"
	WebhookEventQRPhExpired                = "qrph.expired"
)

type WebhookSignature struct {
	Timestamp     int64
	TestSignature string
	LiveSignature string
}

type WebhookEvent struct {
	Data WebhookEventData `json:"data"`
}

type WebhookEventData struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Attributes WebhookEventAttributes `json:"attributes"`
}

type WebhookEventAttributes struct {
	Type            string         `json:"type"`
	Livemode        bool           `json:"livemode"`
	Data            map[string]any `json:"data"`
	PreviousData    map[string]any `json:"previous_data"`
	PendingWebhooks int            `json:"pending_webhooks"`
	CreatedAt       int64          `json:"created_at"`
	UpdatedAt       int64          `json:"updated_at"`
}

type CreateWebhookRequest struct {
	Data CreateWebhookData `json:"data"`
}

type CreateWebhookData struct {
	Attributes CreateWebhookAttributes `json:"attributes"`
}

type CreateWebhookAttributes struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

type CreateWebhookResponse struct {
	Data CreateWebhookResponseData `json:"data"`
}

type CreateWebhookResponseData struct {
	ID         string                          `json:"id"`
	Type       string                          `json:"type"`
	Attributes CreateWebhookResponseAttributes `json:"attributes"`
}

type CreateWebhookResponseAttributes struct {
	Livemode  bool     `json:"livemode"`
	SecretKey string   `json:"secret_key"`
	Status    string   `json:"status"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	CreatedAt int64    `json:"created_at"`
	UpdatedAt int64    `json:"updated_at"`
}

func ParseWebhookSignature(header string) (*WebhookSignature, error) {
	if header == "" {
		return nil, errs.ErrPaymongoWebhookSignatureInvalid
	}

	sig := &WebhookSignature{}
	parts := strings.SplitSeq(header, ",")

	for part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "t":
			ts, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, errors.Join(errs.ErrPaymongoWebhookSignatureInvalid, err)
			}
			sig.Timestamp = ts
		case "te":
			sig.TestSignature = value
		case "li":
			sig.LiveSignature = value
		}
	}

	if sig.TestSignature != "" && sig.LiveSignature != "" {
		return nil, errs.ErrPaymongoWebhookSignatureInvalid
	}

	if sig.Timestamp == 0 {
		return nil, errs.ErrPaymongoWebhookSignatureInvalid
	}

	return sig, nil
}

func VerifyWebhookSignature(payload []byte, signature *WebhookSignature, secretKey string, isLiveMode bool) error {
	if secretKey == "" {
		return errs.ErrPaymongoWebhookSecretRequired
	}

	signatureString := fmt.Sprintf("%d.%s", signature.Timestamp, string(payload))

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(signatureString))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	var actualSignature string
	if isLiveMode {
		actualSignature = signature.LiveSignature
	} else {
		actualSignature = signature.TestSignature
	}

	if !hmac.Equal([]byte(expectedSignature), []byte(actualSignature)) {
		return errs.ErrPaymongoWebhookSignatureInvalid
	}

	return nil
}

func VerifyWebhookTimestamp(timestamp int64, maxAge int64) bool {
	now := time.Now().Unix()
	diff := now - timestamp
	return diff >= 0 && diff <= maxAge
}

func (p *PayMongo) CreateWebhook(webhookURL string, events []string) (*CreateWebhookResponse, error) {
	const logTag = "[PayMongo Create Webhook]"

	payload := CreateWebhookRequest{
		Data: CreateWebhookData{
			Attributes: CreateWebhookAttributes{
				URL:    webhookURL,
				Events: events,
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Join(errs.ErrPaymongoWebhookPayloadInvalid, err)
	}

	URL := p.baseURL + "/webhooks"
	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(jsonPayload))
	if err != nil {
		return nil, errors.Join(errs.ErrPaymentClient, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", p.GetAuth())

	resp, err := p.client.Do(req)
	if err != nil {
		logs.JSONResponse(logTag, resp)
		return nil, errors.Join(errs.ErrPaymentClient, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Log().Error("Deferred", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		logs.JSONResponse(logTag, resp)
		return nil, fmt.Errorf("%w: status code %d", errs.ErrPaymentClient, resp.StatusCode)
	}

	var result CreateWebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logs.JSONResponse(logTag, resp)
		return nil, errors.Join(errs.ErrPaymentResponse, err)
	}

	return &result, nil
}

func IsLiveMode() bool {
	cfg := conf.Conf()
	return strings.HasPrefix(cfg.PayMongo.APIKey, "sk_live_")
}

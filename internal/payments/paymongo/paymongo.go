package paymongo

import (
	"bytes"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/payments"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

type PayMongo struct {
	name   string
	apiKey string
}

func MustInit() *PayMongo {
	apiKey := os.Getenv("PAYMONGO_API_KEY")
	if apiKey == "" {
		panic("PAYMONGO_API_KEY is required")
	}
	apiKey = base64.StdEncoding.EncodeToString([]byte(apiKey))
	return &PayMongo{
		name:   "PayMongo",
		apiKey: apiKey,
	}
}

func (p PayMongo) GatewayName() string {
	return p.name
}

func (p PayMongo) GetAuth() string {
	return "Basic " + p.apiKey
}

func (p PayMongo) CreateCheckoutSession(
	payload payments.CreateCheckoutSessionPayload,
) (payments.CreateCheckoutSessionResponse, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Join(errs.ERR_PAYMENT_PAYLOAD, err)
	}

	const url = "https://api.paymongo.com/v1/checkout_sessions"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonPayload))
	if err != nil {
		return nil, errors.Join(errs.ERR_PAYMENT_CLIENT, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", p.GetAuth())

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil || resp.StatusCode != 200 {
		logs.JSONResponse("[PayMongo] CreateCheckoutSession", resp)
		return nil, errors.Join(errs.ERR_PAYMENT_CLIENT, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Log().Error("Deferred", zap.Error(err))
		}
	}()

	var res CreateCheckoutSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logs.JSONResponse("[PayMongo] CreateCheckoutSession", resp)
		return nil, errors.Join(errs.ERR_PAYMENT_RESPONSE, err)
	}
	return &res, nil
}

func (p PayMongo) GetAvailablePaymentMethods() (payments.GetAvailablePaymentMethodsResponse, error) {
	const url = "https://api.paymongo.com/v1/merchants/capabilities/payment_methods"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Join(errs.ERR_PAYMENT_CLIENT, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", p.GetAuth())

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil || resp.StatusCode != 200 {
		logs.JSONResponse("[PayMongo] GetAvailablePaymentMethods", resp)
		return nil, errors.Join(errs.ERR_PAYMENT_CLIENT, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Log().Error("Deferred", zap.Error(err))
		}
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.JSONResponse("[PayMongo] GetAvailablePaymentMethods", resp)
		return nil, errors.Join(errs.ERR_PAYMENT_RESPONSE, err)
	}

	var res GetAvailablePaymentMethodsResponse
	if err := json.Unmarshal(data, &res.Data); err != nil {
		logs.JSONResponse("[PayMongo] GetAvailablePaymentMethods", resp)
		return nil, errors.Join(errs.ERR_PAYMENT_RESPONSE, err)
	}
	return &res, nil
}

var _ payments.PaymentGateway = (*PayMongo)(nil)

package payments

import (
	"bytes"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"encoding/base64"
	"errors"
	"net/http"
	"os"

	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

type PayMongo struct {
	name   string
	apiKey string
}

func MustInitPayMongo() *PayMongo {
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

func (p PayMongo) CreateCheckoutSession(payload createCheckoutSessionPayload) (createCheckoutSessionResponse, error) {
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
	if err != nil || resp == nil {
		return nil, errors.Join(errs.ERR_PAYMENT_CLIENT, err)
	}
	logs.LogResBody(logs.Log(), "PayMongoCreateCheckoutSessionResponse", resp)

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Log().Error("Deferred", zap.Error(err))
		}
	}()
	var res PayMongoCreateCheckoutSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.Join(errs.ERR_PAYMENT_RESPONSE, err)
	}
	return &res, nil
}

var _ IPayments = (*PayMongo)(nil)

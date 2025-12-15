package paymongo

import (
	"bytes"
	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/payments"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gookit/goutil/dump"
	"go.uber.org/zap"
)

type PayMongo struct {
	client         *http.Client
	apiKey         string
	successURL     string
	cancelURL      string
	baseURL        string
	paymentGateway payments.PaymentGateway
}

func validate() {
	cfg := conf.Conf()
	if cfg.PaymentService != payments.PAYMENT_GATEWAY_PAYMONGO.String() {
		panic(errs.ErrPaymongoServiceInit)
	}
	if cfg.PayMongo.BaseURL == "" || cfg.PayMongo.APIKey == "" || cfg.PayMongo.SuccessURL == "" || cfg.PayMongo.CancelURL == "" {
		panic(errs.ErrPaymongoAPIKeyRequired)
	}

	if cfg.IsProd() {
		if !strings.HasPrefix(cfg.PayMongo.APIKey, "sk_live_") {
			panic(fmt.Errorf("%w: production environment requires 'sk_live_' API key", errs.ErrPaymongoAPIKeyInvalid))
		}
	} else {
		if !strings.HasPrefix(cfg.PayMongo.APIKey, "sk_test_") {
			panic(fmt.Errorf("%w: non-production environment requires 'sk_test_' API key", errs.ErrPaymongoAPIKeyInvalid))
		}
	}
}

func MustInit() *PayMongo {
	validate()

	cfg := conf.Conf()
	apiKey := base64.StdEncoding.EncodeToString([]byte(cfg.PayMongo.APIKey))

	var successURL, cancelURL string
	if cfg.Server.Address == "localhost" {
		successURL = fmt.Sprintf("http://localhost:%d%s", cfg.Server.Port, cfg.PayMongo.SuccessURL)
		cancelURL = fmt.Sprintf("http://localhost:%d%s", cfg.Server.Port, cfg.PayMongo.CancelURL)
	} else {
		successURL = cfg.PayMongo.SuccessURL
		cancelURL = cfg.PayMongo.CancelURL
	}

	return &PayMongo{
		paymentGateway: payments.PAYMENT_GATEWAY_PAYMONGO,
		apiKey:         apiKey,
		successURL:     successURL,
		cancelURL:      cancelURL,
		baseURL:        cfg.PayMongo.BaseURL,
		client:         &http.Client{Timeout: 10 * time.Second},
	}
}

func (p PayMongo) GatewayEnum() payments.PaymentGateway {
	return p.paymentGateway
}

func (p PayMongo) GetAuth() string {
	return "Basic " + p.apiKey
}

func (p PayMongo) CreateCheckoutPaymentSession(
	payload payments.CreateCheckoutSessionPayload,
) (payments.CreateCheckoutSessionResponse, error) {
	const logTag = "[PayMongo Create Checkout Payment Session]"
	paymongoPayload, ok := payload.(CreateCheckoutSessionPayload)
	if !ok {
		return nil, fmt.Errorf("%w. Not for PayMongo", errs.ErrPaymentPayload)
	}

	jsonPayload, err := json.Marshal(paymongoPayload)
	if err != nil {
		return nil, errors.Join(errs.ErrPaymentPayload, err)
	}

	URL := p.baseURL + "/checkout_sessions"
	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(jsonPayload))
	if err != nil {
		return nil, errors.Join(errs.ErrPaymentClient, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", p.GetAuth())

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		logs.JSONResponse(logTag, resp)
		return nil, errors.Join(errs.ErrPaymentClient, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Log().Error("Deferred", zap.Error(err))
		}
	}()

	var res CreateCheckoutSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logs.JSONResponse(logTag, resp)
		return nil, errors.Join(errs.ErrPaymentResponse, err)
	}
	return &res, nil
}

func (p PayMongo) GetAvailablePaymentMethods() (payments.GetAvailablePaymentMethodsResponse, error) {
	const logTag = "[PayMongo Get Available Payment Methods]"
	URL := p.baseURL + "/merchants/capabilities/payment_methods"
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, errors.Join(errs.ErrPaymentClient, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", p.GetAuth())

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		logs.JSONResponse(logTag, resp)
		return nil, errors.Join(errs.ErrPaymentClient, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Log().Error("Deferred", zap.Error(err))
		}
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.JSONResponse(logTag, resp)
		return nil, errors.Join(errs.ErrPaymentResponse, err)
	}

	var res GetAvailablePaymentMethodsResponse
	if err := json.Unmarshal(data, &res.Data); err != nil {
		logs.JSONResponse(logTag, resp)
		return nil, errors.Join(errs.ErrPaymentResponse, err)
	}
	return &res, nil
}

func (p PayMongo) CheckoutPaymentHandler(
	w http.ResponseWriter,
	r *http.Request,
	payload payments.CreateCheckoutSessionPayload,
) error {
	resCheckout, err := p.CreateCheckoutPaymentSession(payload)
	if err != nil {
		return err
	}

	resPayMongoCheckout := resCheckout.(*CreateCheckoutSessionResponse)
	w.Header().Set("HX-Redirect", resPayMongoCheckout.Data.Attributes.CheckoutURL)

	return nil
}

// INFO: (Brandon) - Format: CC{gatewayenum}-{time}{upper 6 chars}
func (p PayMongo) GenerateRefNo() string {
	gatewayCode := p.GatewayEnum().Code()

	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		ts := time.Now().UTC().UnixNano()
		return strings.ToUpper(fmt.Sprintf("CC%s-%06x%d", gatewayCode, ts%0xFFFFFF, ts))
	}

	randomChars := hex.EncodeToString(b)
	ts := time.Now().UTC().UnixMilli()
	tsStr := strconv.FormatInt(ts, 36)
	return strings.ToUpper(fmt.Sprintf("CC%s-%s%s", gatewayCode, tsStr, randomChars))
}

func (p PayMongo) CreatePayload(
	billing payments.Billing,
	lineItems []payments.LineItem,
	paymentMethods []payments.PaymentMethod,
) payments.CreateCheckoutSessionPayload {
	referenceNumber := p.GenerateRefNo()

	cancelURLWithRef := p.cancelURL
	if u, err := url.Parse(cancelURLWithRef); err == nil {
		q := u.Query()
		q.Set("payment_ref", referenceNumber)
		u.RawQuery = q.Encode()
		cancelURLWithRef = u.String()
	}

	successURLWithRef := p.successURL
	if u, err := url.Parse(successURLWithRef); err == nil {
		q := u.Query()
		q.Set("payment_ref", referenceNumber)
		u.RawQuery = q.Encode()
		successURLWithRef = u.String()
	}

	billing.Phone = strings.TrimPrefix(billing.Phone, "+63")
	paymentMethodNames := make([]string, 0, len(paymentMethods))
	for _, pm := range paymentMethods {
		paymentMethodNames = append(paymentMethodNames, strings.ToLower(pm.String()))
	}

	payload := CreateCheckoutSessionPayload{
		Data: CreateCheckoutSessionData{
			Attributes: CreateCheckoutSessionAttr{
				CancelURL:           cancelURLWithRef,
				SuccessURL:          successURLWithRef,
				Billing:             billing,
				LineItems:           lineItems,
				PaymentMethodTypes:  paymentMethodNames,
				Description:         "C-Choice Checkout",
				ReferenceNumber:     referenceNumber,
				SendEmailReceipt:    conf.Conf().IsProd(),
				ShowDescription:     true,
				ShowLineItems:       true,
				StatementDescriptor: "C-Choice Checkout Statement",
			},
		},
	}
	if conf.Conf().IsLocal() {
		dump.Println("PAYMONGO PAYLOAD", payload)
	}
	return payload
}

func (p PayMongo) GetPaymentIntent(paymentIntentID string) (*GetPaymentIntentResponse, error) {
	const logTag = "[PayMongo Get Payment Intent]"
	URL := fmt.Sprintf("%s/payment_intents/%s", p.baseURL, paymentIntentID)
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, errors.Join(errs.ErrPaymentClient, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", p.GetAuth())

	resp, err := p.client.Do(req)
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		logs.JSONResponse(logTag, resp)
		return nil, errors.Join(errs.ErrPaymentClient, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Log().Error("Deferred", zap.Error(err))
		}
	}()

	var res GetPaymentIntentResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logs.JSONResponse(logTag, resp)
		return nil, errors.Join(errs.ErrPaymentResponse, err)
	}
	return &res, nil
}

var _ payments.IPaymentGateway = (*PayMongo)(nil)

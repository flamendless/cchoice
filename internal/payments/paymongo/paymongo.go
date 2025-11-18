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

func MustInit() *PayMongo {
	cfg := conf.Conf()
	if cfg.PaymentService != "paymongo" {
		panic(errs.ErrPaymongoServiceInit)
	}

	apiKey := base64.StdEncoding.EncodeToString([]byte(cfg.PayMongo.APIKey))

	address := cfg.Server.Address
	if address == "localhost" {
		address = fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
	}

	successURL := fmt.Sprintf("%s%s", address, cfg.PayMongo.SuccessURL)
	cancelURL := fmt.Sprintf("%s%s", address, cfg.PayMongo.CancelURL)

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
		logs.JSONResponse("[PayMongo] CreateCheckoutSession", resp)
		return nil, errors.Join(errs.ErrPaymentClient, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Log().Error("Deferred", zap.Error(err))
		}
	}()

	var res CreateCheckoutSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logs.JSONResponse("[PayMongo] CreateCheckoutSession", resp)
		return nil, errors.Join(errs.ErrPaymentResponse, err)
	}
	return &res, nil
}

func (p PayMongo) GetAvailablePaymentMethods() (payments.GetAvailablePaymentMethodsResponse, error) {
	URL := p.baseURL + "/merchants/capabilities/payment_methods"
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, errors.Join(errs.ErrPaymentClient, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", p.GetAuth())

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		logs.JSONResponse("[PayMongo] GetAvailablePaymentMethods", resp)
		return nil, errors.Join(errs.ErrPaymentClient, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Log().Error("Deferred", zap.Error(err))
		}
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.JSONResponse("[PayMongo] GetAvailablePaymentMethods", resp)
		return nil, errors.Join(errs.ErrPaymentResponse, err)
	}

	var res GetAvailablePaymentMethodsResponse
	if err := json.Unmarshal(data, &res.Data); err != nil {
		logs.JSONResponse("[PayMongo] GetAvailablePaymentMethods", resp)
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
		q.Set("order_ref", referenceNumber)
		u.RawQuery = q.Encode()
		cancelURLWithRef = u.String()
	}

	successURLWithRef := p.successURL
	if u, err := url.Parse(successURLWithRef); err == nil {
		q := u.Query()
		q.Set("order_ref", referenceNumber)
		u.RawQuery = q.Encode()
		successURLWithRef = u.String()
	}

	payload := CreateCheckoutSessionPayload{
		Data: CreateCheckoutSessionData{
			Attributes: CreateCheckoutSessionAttr{
				CancelURL:           cancelURLWithRef,
				SuccessURL:          successURLWithRef,
				Billing:             billing,
				LineItems:           lineItems,
				PaymentMethodTypes:  paymentMethods,
				Description:         "C-Choice Checkout",
				ReferenceNumber:     referenceNumber,
				SendEmailReceipt:    false,
				ShowDescription:     true,
				ShowLineItems:       true,
				StatementDescriptor: "C-Choice Checkout Statement",
			},
		},
	}
	return payload
}

var _ payments.IPaymentGateway = (*PayMongo)(nil)

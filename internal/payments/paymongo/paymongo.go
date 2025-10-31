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
		panic("'PAYMENT_SERVICE' must be 'paymongo' to use this")
	}

	apiKey := base64.StdEncoding.EncodeToString([]byte(cfg.PayMongoAPIKey))

	address := cfg.Address
	if address == "localhost" {
		address = "http://localhost"
	}
	successURL := fmt.Sprintf("%s:%d%s", address, cfg.Port, cfg.PayMongoSuccessURL)
	cancelURL := fmt.Sprintf("%s:%d%s", address, cfg.Port, cfg.PayMongoCancelURL)

	return &PayMongo{
		paymentGateway: payments.PAYMENT_GATEWAY_PAYMONGO,
		apiKey:         apiKey,
		successURL:     successURL,
		cancelURL:      cancelURL,
		baseURL:        cfg.PayMongoBaseURL,
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

func (p PayMongo) GenerateRefNo() string {
	ts := time.Now().UTC().UnixNano()
	tsEnc := strconv.FormatInt(ts, 36)

	//INFO: (Brandon) - if we ever encounter duplication, increase precision by increasing number of bytes
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%s_%s_%d", p.GatewayEnum().Code(), tsEnc, time.Now().UnixNano())
	}

	r := hex.EncodeToString(b)
	return fmt.Sprintf("%s_%s_%s", p.GatewayEnum().Code(), tsEnc, r)
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

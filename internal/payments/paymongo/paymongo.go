package paymongo

import (
	"bytes"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/payments"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Rhymond/go-money"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

type PayMongo struct {
	apiKey         string
	successURL     string
	paymentGateway payments.PaymentGateway
}

func MustInit() *PayMongo {
	apiKey := os.Getenv("PAYMONGO_API_KEY")
	if apiKey == "" {
		panic(fmt.Errorf("%w. PAYMONGO_API_KEY", errs.ERR_ENV_VAR_REQUIRED))
	}
	apiKey = base64.StdEncoding.EncodeToString([]byte(apiKey))

	successURL := os.Getenv("PAYMONGO_SUCCESS_URL")
	if successURL == "" {
		panic(fmt.Errorf("%w. PAYMONGO_SUCCESS_URL", errs.ERR_ENV_VAR_REQUIRED))
	}

	return &PayMongo{
		paymentGateway: payments.PAYMENT_GATEWAY_PAYMONGO,
		apiKey:         apiKey,
		successURL:     successURL,
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
	if _, ok := payload.(*CreateCheckoutSessionPayload); !ok {
		return nil, fmt.Errorf("%w. Not for PayMongo", errs.ERR_PAYMENT_PAYLOAD)
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Join(errs.ERR_PAYMENT_PAYLOAD, err)
	}

	const URL = "https://api.paymongo.com/v1/checkout_sessions"
	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(jsonPayload))
	if err != nil {
		return nil, errors.Join(errs.ERR_PAYMENT_CLIENT, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", p.GetAuth())

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
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
	const URL = "https://api.paymongo.com/v1/merchants/capabilities/payment_methods"
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, errors.Join(errs.ERR_PAYMENT_CLIENT, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", p.GetAuth())

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
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

func (p PayMongo) CheckoutPaymentHandler(w http.ResponseWriter, r *http.Request) error {
	resPaymentMethods, err := p.GetAvailablePaymentMethods()
	if err != nil {
		return err
	}

	payload := CreateCheckoutSessionPayload{
		Data: CreateCheckoutSessionData{
			Attributes: CreateCheckoutSessionAttr{
				CancelURL:  "https://test.com/cancel",
				SuccessURL: p.successURL,
				Billing: payments.Billing{
					Address: payments.Address{
						Line1:      "test line 1",
						Line2:      "test line 2",
						City:       "test city",
						State:      "test state",
						PostalCode: "test postal code",
						Country:    "PH",
					},
					Name:  "test name",
					Email: "test@mail.com",
					Phone: "test phone",
				},
				LineItems: []payments.LineItem{
					{
						Amount:      1000,
						Currency:    money.PHP,
						Description: "test line item description",
						Images:      []string{"https://test.com/image"},
						Name:        "test line item name",
						Quantity:    2,
					},
				},
				Description:         "test description",
				PaymentMethodTypes:  resPaymentMethods.ToPaymentMethods(),
				ReferenceNumber:     "test-ref-number",
				SendEmailReceipt:    false,
				ShowDescription:     true,
				ShowLineItems:       true,
				StatementDescriptor: "test statement descriptor",
			},
		},
	}

	resCheckout, err := p.CreateCheckoutPaymentSession(&payload)
	if err != nil {
		return err
	}

	resPayMongoCheckout := resCheckout.(*CreateCheckoutSessionResponse)
	w.Header().Set("HX-Redirect", resPayMongoCheckout.Data.Attributes.CheckoutURL)

	return nil
}

var _ payments.IPaymentGateway = (*PayMongo)(nil)

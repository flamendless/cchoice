package lalamove

import (
	"bytes"
	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"cchoice/internal/shipping"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Lalamove struct {
	client           *http.Client
	apiKey           string
	secret           string
	baseURL          string
	apiVersion       string
	shippingService  shipping.ShippingService
	businessLocation *shipping.Location
}

func MustInit() *Lalamove {
	cfg := conf.Conf()
	if cfg.ShippingService != "lalamove" {
		panic(errs.ErrLalamoveServiceInit)
	}

	return &Lalamove{
		shippingService: shipping.SHIPPING_SERVICE_LALAMOVE,
		apiKey:          cfg.LalamoveAPIKey,
		secret:          cfg.LalamoveAPISecret,
		baseURL:         cfg.LalamoveBaseURL,
		client:          &http.Client{Timeout: 10 * time.Second},
		apiVersion:      "v3",
		businessLocation: &shipping.Location{
			Coordinates: shipping.Coordinates{
				Lat: cfg.BusinessLat,
				Lng: cfg.BusinessLng,
			},
			Address: cfg.BusinessAddress,
		},
	}
}

func (c *Lalamove) Enum() shipping.ShippingService {
	return c.shippingService
}

func (c *Lalamove) GetBusinessLocation() *shipping.Location {
	return c.businessLocation
}

func (c *Lalamove) signRequest(method string, path string, body []byte) (string, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	message := fmt.Sprintf("%s\r\n%s\r\n%s\r\n\r\n%s", timestamp, method, path, string(body))

	mac := hmac.New(sha256.New, []byte(c.secret))
	if _, err := mac.Write([]byte(message)); err != nil {
		return "", errors.Join(errs.ErrLalamoveSignRequest, err)
	}
	signature := hex.EncodeToString(mac.Sum(nil))
	authHeader := fmt.Sprintf("hmac %s:%s:%s", c.apiKey, timestamp, signature)
	return authHeader, nil
}

func (c *Lalamove) doRequest(method, path string, body []byte) (*http.Response, error) {
	if len(body) > 0 {
		wrappedBody := map[string]json.RawMessage{
			"data": json.RawMessage(body),
		}
		rawWrappedBody, err := json.Marshal(wrappedBody)
		if err != nil {
			return nil, errors.Join(errs.ErrLalamoveJSONMarshal, errs.ErrJSONMarshal, err)
		}
		body = rawWrappedBody
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Join(errs.ErrLalamoveHTTPRequest, errs.ErrHTTPNewRequest, err)
	}

	authHeader, err := c.signRequest(method, path, body)
	if err != nil {
		return nil, errors.Join(errs.ErrLalamoveSignRequest, err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Market", "PH")

	resp, err := c.client.Do(req)
	if err != nil || resp == nil {
		return nil, errors.Join(errs.ErrLalamoveHTTPRequest, errs.ErrHTTPDoRequest, err)
	}

	if resp.ContentLength != 0 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Join(errs.ErrLalamoveHTTPRequest, errs.ErrHTTPReadResponse, err)
		}
		resp.Body.Close()

		var wrapper map[string]json.RawMessage
		if err := json.Unmarshal(bodyBytes, &wrapper); err != nil {
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			return resp, errors.Join(errs.ErrLalamoveJSONUnmarshal, errs.ErrJSONUnmarshal, err)
		}

		if errorsValue, hasErrors := wrapper["errors"]; hasErrors {
			var errorResp LalamoveErrorResponse
			if err := json.Unmarshal(errorsValue, &errorResp.Errors); err == nil {
				return nil, errors.Join(errs.ErrLalamoveAPIResponse, errorResp)
			} else {
				return nil, errors.Join(errs.ErrLalamoveJSONUnmarshal, errs.ErrJSONUnmarshal, err)
			}
		}

		if messageValue, hasMessage := wrapper["message"]; hasMessage {
			var message string
			if err := json.Unmarshal(messageValue, &message); err == nil {
				return nil, errs.ErrLalamoveAPIResponse
			} else {
				return nil, errors.Join(errs.ErrLalamoveJSONUnmarshal, errs.ErrJSONUnmarshal, err)
			}
		}

		if resp.StatusCode < 300 {
			if dataValue, exists := wrapper["data"]; exists {
				resp.Body = io.NopCloser(bytes.NewReader(dataValue))
				resp.ContentLength = int64(len(dataValue))
				return resp, nil
			}
		}

		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	return resp, nil
}

func (c *Lalamove) GetQuotation(req shipping.ShippingRequest) (*shipping.ShippingQuotation, error) {
	lalamoveReq := NewLalamoveQuotationRequest(req)
	body, err := json.Marshal(lalamoveReq)
	if err != nil {
		return nil, errors.Join(errs.ErrLalamoveQuotation, errs.ErrLalamoveJSONMarshal, err)
	}
	resp, err := c.doRequest(http.MethodPost, "/v3/quotations", body)
	if err != nil || resp == nil {
		return nil, errors.Join(errs.ErrLalamoveQuotation, err)
	}
	defer resp.Body.Close()

	var result QuotationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Join(errs.ErrLalamoveQuotation, errs.ErrJSONDecode, err)
	}

	return result.ToShippingQuotation(), nil
}

func (c *Lalamove) CreateOrder(req shipping.ShippingRequest) (*shipping.ShippingOrder, error) {
	lalamoveReq := NewLalamoveOrderRequest(req)
	body, err := json.Marshal(lalamoveReq)
	if err != nil {
		return nil, errors.Join(errs.ErrLalamoveOrderCreate, errs.ErrLalamoveJSONMarshal, err)
	}

	resp, err := c.doRequest(http.MethodPost, "/v3/orders", body)
	if err != nil || resp == nil {
		return nil, errors.Join(errs.ErrLalamoveOrderCreate, err)
	}
	defer resp.Body.Close()

	var result OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Join(errs.ErrLalamoveOrderCreate, errs.ErrJSONDecode, err)
	}

	return result.ToShippingOrder(), nil
}

func (c *Lalamove) GetOrderStatus(orderID string) (*shipping.ShippingOrder, error) {
	path := "/v3/orders/" + orderID

	resp, err := c.doRequest("GET", path, nil)
	if err != nil || resp == nil {
		return nil, errors.Join(errs.ErrLalamoveOrderStatus, err)
	}
	defer resp.Body.Close()

	var result OrderStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Join(errs.ErrLalamoveOrderStatus, errs.ErrJSONDecode, err)
	}

	return result.ToShippingOrder(), nil
}

func (c *Lalamove) CancelOrder(orderID string) error {
	path := fmt.Sprintf("/v3/orders/%s/cancel", orderID)

	resp, err := c.doRequest("PUT", path, nil)
	if err != nil || resp == nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *Lalamove) GetCapabilities() (*shipping.ServiceCapabilities, error) {
	resp, err := c.doRequest("GET", "/v3/cities", nil)
	if err != nil || resp == nil {
		return nil, errors.Join(errs.ErrLalamoveCapabilities, err)
	}
	defer resp.Body.Close()

	var cities CitiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&cities); err != nil {
		return nil, errors.Join(errs.ErrLalamoveCapabilities, errs.ErrJSONDecode, err)
	}

	coverage := make([]string, 0, len(cities))
	for _, city := range cities {
		coverage = append(coverage, city.Name)
	}

	var supportedServices []shipping.ServiceType
	serviceTypeSet := make(map[shipping.ServiceType]bool)
	for _, city := range cities {
		for _, service := range city.Services {
			serviceType := shipping.ParseServiceTypeToEnum(service.Key)
			if !serviceTypeSet[serviceType] {
				supportedServices = append(supportedServices, serviceType)
				serviceTypeSet[serviceType] = true
			}
		}
	}

	return &shipping.ServiceCapabilities{
		SupportedServices: supportedServices,
		Coverage:          coverage,
		Features: shipping.Features{
			RealTimeTracking:    true,
			RouteOptimization:   true,
			ScheduledDelivery:   true,
			SpecialRequests:     true,
			MultipleStops:       true,
			WeightBasedPricing:  true,
			Insurance:           false,
			ProofOfDelivery:     true,
			CashOnDelivery:      true,
			ContactlessDelivery: true,
		},
		Provider:   c.shippingService.String(),
		APIVersion: c.apiVersion,
		Metadata: map[string]any{
			"cities_data": cities,
		},
	}, nil
}

var _ shipping.IShippingService = (*Lalamove)(nil)

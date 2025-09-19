package lalamove

import (
	"cchoice/internal/shipping"
	"encoding/json"
	"strconv"
)

type DimensionValue struct {
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

type Dimensions struct {
	Length DimensionValue `json:"length"`
	Width  DimensionValue `json:"width"`
	Height DimensionValue `json:"height"`
}

type LoadValue struct {
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

type Service struct {
	Key             string           `json:"key"`
	Description     string           `json:"description"`
	Dimensions      Dimensions       `json:"dimensions"`
	Load            LoadValue        `json:"load"`
	SpecialRequests []SpecialRequest `json:"specialRequests"`
}

type SpecialRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	ParentType    string `json:"parent_type,omitempty"`
	MaxSelection  int    `json:"max_selection"`
	EffectiveTime string `json:"effective_time,omitempty"`
	OfflineTime   string `json:"offline_time,omitempty"`
}

type City struct {
	Name     string    `json:"name"`
	Locode   string    `json:"locode"`
	Services []Service `json:"services"`
}

type CitiesResponse []City

type Price float64

func (p *Price) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}

	*p = Price(val)
	return nil
}

func (p Price) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatFloat(float64(p), 'f', -1, 64))
}

type DistanceValue struct {
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

type DistanceKm float64

func (d *DistanceKm) UnmarshalJSON(data []byte) error {
	var distanceStruct DistanceValue

	if err := json.Unmarshal(data, &distanceStruct); err != nil {
		return err
	}

	meters, err := strconv.ParseFloat(distanceStruct.Value, 64)
	if err != nil {
		return err
	}

	*d = DistanceKm(meters / 1000.0)
	return nil
}

func (d DistanceKm) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"value": strconv.FormatFloat(float64(d)*1000, 'f', -1, 64),
		"unit":  "m",
	})
}

type QuotationResponse struct {
	QuotationID      string                  `json:"quotationId"`
	ScheduleAt       string                  `json:"scheduleAt"`
	ExpiresAt        string                  `json:"expiresAt"`
	ServiceType      string                  `json:"serviceType"`
	SpecialRequests  []string                `json:"specialRequests"`
	Language         string                  `json:"language"`
	Stops            []QuotationStop         `json:"stops"`
	IsRouteOptimized bool                    `json:"isRouteOptimized"`
	PriceBreakdown   QuotationPriceBreakdown `json:"priceBreakdown"`
	Item             QuotationItem           `json:"item"`
	Distance         DistanceKm              `json:"distance"`
}

type QuotationStop struct {
	StopID      string               `json:"stopId"`
	Coordinates shipping.Coordinates `json:"coordinates"`
	Address     string               `json:"address"`
}

type QuotationPriceBreakdown struct {
	Base                    Price  `json:"base"`
	SpecialRequests         Price  `json:"specialRequests"`
	VAT                     Price  `json:"vat"`
	TotalBeforeOptimization Price  `json:"totalBeforeOptimization"`
	TotalExcludePriorityFee Price  `json:"totalExcludePriorityFee"`
	Total                   Price  `json:"total"`
	Currency                string `json:"currency"`
}

type QuotationItem struct {
	Weight     string   `json:"weight"`
	Categories []string `json:"categories"`
}

type OrderResponse struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
}

type OrderStatusResponse struct {
	OrderID  string  `json:"orderId"`
	Status   string  `json:"status"`
	Currency string  `json:"currency"`
	Price    Price   `json:"price"`
	Distance float64 `json:"distance"`
	ETA      int     `json:"time"`
}

func (q *QuotationResponse) ToShippingQuotation() *shipping.ShippingQuotation {
	return &shipping.ShippingQuotation{
		ID:           q.QuotationID,
		Currency:     q.PriceBreakdown.Currency,
		Fee:          float64(q.PriceBreakdown.Total),
		DistanceKm:   float64(q.Distance),
		EstimatedETA: 0,
		ServiceType:  q.ServiceType,
		ExpiresAt:    q.ExpiresAt,
		Metadata: map[string]any{
			"schedule_at":        q.ScheduleAt,
			"special_requests":   q.SpecialRequests,
			"language":           q.Language,
			"is_route_optimized": q.IsRouteOptimized,
			"price_breakdown": map[string]any{
				"base":                       float64(q.PriceBreakdown.Base),
				"special_requests":           float64(q.PriceBreakdown.SpecialRequests),
				"vat":                        float64(q.PriceBreakdown.VAT),
				"total_before_optimization":  float64(q.PriceBreakdown.TotalBeforeOptimization),
				"total_exclude_priority_fee": float64(q.PriceBreakdown.TotalExcludePriorityFee),
			},
		},
	}
}

func (o *OrderResponse) ToShippingOrder() *shipping.ShippingOrder {
	return &shipping.ShippingOrder{
		ID:     o.OrderID,
		Status: o.Status,
		Metadata: map[string]any{
			"created_at": "unknown",
		},
	}
}

func (o *OrderStatusResponse) ToShippingOrder() *shipping.ShippingOrder {
	return &shipping.ShippingOrder{
		ID:     o.OrderID,
		Status: o.Status,
		Quotation: shipping.ShippingQuotation{
			Currency:     o.Currency,
			Fee:          float64(o.Price),
			DistanceKm:   o.Distance,
			EstimatedETA: o.ETA,
		},
		Metadata: map[string]any{
			"last_updated": "unknown",
		},
	}
}

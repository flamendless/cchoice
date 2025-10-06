package lalamove

import (
	"cchoice/internal/shipping"
	"maps"
)

type QuotationStopRequest struct {
	Coordinates shipping.Coordinates `json:"coordinates"`
	Address     string               `json:"address"`
}

type QuotationItemRequest struct {
	Weight     string   `json:"weight"`
	Categories []string `json:"categories"`
}

type LalamoveQuotationRequest struct {
	ScheduleAt       string                 `json:"scheduleAt,omitempty"`
	ServiceType      string                 `json:"serviceType"`
	SpecialRequests  []string               `json:"specialRequests,omitempty"`
	Language         string                 `json:"language"`
	Stops            []QuotationStopRequest `json:"stops"`
	Item             QuotationItemRequest   `json:"item"`
	IsRouteOptimized bool                   `json:"isRouteOptimized,omitempty"`
}

type OrderSender struct {
	StopID string `json:"stopId"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
}

type OrderRecipient struct {
	StopID  string `json:"stopId"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Remarks string `json:"remarks,omitempty"`
}

type OrderMetadata map[string]string

type LalamoveOrderRequest struct {
	IsPODEnabled *bool            `json:"isPODEnabled,omitempty"`
	Metadata     *OrderMetadata   `json:"metadata,omitempty"`
	Sender       OrderSender      `json:"sender"`
	QuotationID  string           `json:"quotationId"`
	Partner      string           `json:"partner,omitempty"`
	Recipients   []OrderRecipient `json:"recipients"`
}

func NewLalamoveOrderRequest(req shipping.ShippingRequest) *LalamoveOrderRequest {
	quotationID := ""
	if req.Options != nil {
		if quotationIDVal, exists := req.Options["quotation_id"]; exists {
			if quotationIDStr, ok := quotationIDVal.(string); ok {
				quotationID = quotationIDStr
			}
		}
	}

	var senderStopID, recipientStopID string
	if req.Options != nil {
		if stopsData, exists := req.Options["quotation_stops"]; exists {
			if stops, ok := stopsData.([]QuotationStop); ok {
				if len(stops) >= 1 {
					senderStopID = stops[0].StopID
				}
				if len(stops) >= 2 {
					recipientStopID = stops[1].StopID
				}
			}
		}
	}

	sender := OrderSender{
		StopID: senderStopID,
		Name:   req.PickupLocation.Contact.Name,
		Phone:  req.PickupLocation.Contact.Phone,
	}

	recipients := []OrderRecipient{
		{
			StopID: recipientStopID,
			Name:   req.DeliveryLocation.Contact.Name,
			Phone:  req.DeliveryLocation.Contact.Phone,
		},
	}

	var isPODEnabled *bool
	var partner string
	var metadata *OrderMetadata

	if req.Options != nil {
		if podVal, exists := req.Options["is_pod_enabled"]; exists {
			if podBool, ok := podVal.(bool); ok {
				isPODEnabled = &podBool
			}
		}

		if partnerVal, exists := req.Options["partner"]; exists {
			if partnerStr, ok := partnerVal.(string); ok {
				partner = partnerStr
			}
		}

		if metadataVal, exists := req.Options["metadata"]; exists {
			if metadataMap, ok := metadataVal.(map[string]any); ok {
				metadata = &OrderMetadata{}
				*metadata = make(map[string]string)
				for key, value := range metadataMap {
					if valueStr, ok := value.(string); ok {
						(*metadata)[key] = valueStr
					}
				}
			}
		}

		if remarksVal, exists := req.Options["remarks"]; exists {
			if remarksStr, ok := remarksVal.(string); ok {
				recipients[0].Remarks = remarksStr
			}
		}
	}

	return &LalamoveOrderRequest{
		QuotationID:  quotationID,
		Sender:       sender,
		Recipients:   recipients,
		IsPODEnabled: isPODEnabled,
		Partner:      partner,
		Metadata:     metadata,
	}
}

type OrderRequestParams struct {
	Metadata     map[string]string `json:"metadata,omitempty"`
	Partner      string            `json:"partner,omitempty"`
	Remarks      string            `json:"remarks,omitempty"`
	IsPODEnabled bool              `json:"isPODEnabled"`
}

func CreateOrderRequest(originalReq shipping.ShippingRequest, quotation *shipping.ShippingQuotation, params OrderRequestParams) shipping.ShippingRequest {
	orderReq := shipping.ShippingRequest{
		Package:          originalReq.Package,
		PickupLocation:   originalReq.PickupLocation,
		DeliveryLocation: originalReq.DeliveryLocation,
		ScheduledAt:      originalReq.ScheduledAt,
		ServiceType:      originalReq.ServiceType,
		Options:          make(map[string]any),
	}

	if originalReq.Options != nil {
		maps.Copy(orderReq.Options, originalReq.Options)
	}

	orderReq.Options["quotation_id"] = quotation.ID
	orderReq.Options["is_pod_enabled"] = params.IsPODEnabled

	if quotation.Metadata != nil {
		if stopsData, exists := quotation.Metadata["stops"]; exists {
			orderReq.Options["quotation_stops"] = stopsData
		}
	}

	if params.Partner != "" {
		orderReq.Options["partner"] = params.Partner
	}

	if params.Remarks != "" {
		orderReq.Options["remarks"] = params.Remarks
	}

	if len(params.Metadata) > 0 {
		metadataInterface := make(map[string]any)
		for k, v := range params.Metadata {
			metadataInterface[k] = v
		}
		orderReq.Options["metadata"] = metadataInterface
	}

	return orderReq
}

func NewLalamoveQuotationRequest(req shipping.ShippingRequest) *LalamoveQuotationRequest {
	stops := []QuotationStopRequest{
		{
			Coordinates: req.PickupLocation.Coordinates,
			Address:     req.PickupLocation.Address,
		},
		{
			Coordinates: req.DeliveryLocation.Coordinates,
			Address:     req.DeliveryLocation.Address,
		},
	}

	var categories []string
	if categoriesVal, exists := req.Package.Metadata["categories"]; exists {
		if categoriesSlice, ok := categoriesVal.([]string); ok {
			categories = categoriesSlice
		}
	}

	item := QuotationItemRequest{
		Weight:     req.Package.Weight,
		Categories: categories,
	}

	language := "en_PH"
	var specialRequests []string
	var isRouteOptimized bool

	if req.Options != nil {
		if specialReqVal, exists := req.Options["special_requests"]; exists {
			if specialReqSlice, ok := specialReqVal.([]string); ok {
				specialRequests = specialReqSlice
			}
		}
		if langVal, exists := req.Options["language"]; exists {
			if langStr, ok := langVal.(string); ok {
				language = langStr
			}
		}
		if routeOptVal, exists := req.Options["is_route_optimized"]; exists {
			if routeOptBool, ok := routeOptVal.(bool); ok {
				isRouteOptimized = routeOptBool
			}
		}
	}

	return &LalamoveQuotationRequest{
		ScheduleAt:       req.ScheduledAt,
		ServiceType:      req.ServiceType.String(),
		SpecialRequests:  specialRequests,
		Language:         language,
		Stops:            stops,
		Item:             item,
		IsRouteOptimized: isRouteOptimized,
	}
}

package lalamove

import (
	"cchoice/internal/shipping"
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
		ServiceType:      req.ServiceType,
		SpecialRequests:  specialRequests,
		Language:         language,
		Stops:            stops,
		Item:             item,
		IsRouteOptimized: isRouteOptimized,
	}
}

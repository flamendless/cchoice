package cchoice

import (
	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"cchoice/internal/shipping"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

type InternalService struct {
	shippingService shipping.ShippingService
	baseFee         float64
	feePerKm        float64
	feePerKg        float64
	maxDistance     float64
}

func MustInit() *InternalService {
	cfg := conf.Conf()
	if cfg.ShippingService != "cchoice" {
		panic("'SHIPPING_SERVICE' must be 'cchoice' to use this")
	}

	return &InternalService{
		shippingService: shipping.SHIPPING_SERVICE_CCHOICE,
		baseFee:         50.0,
		feePerKm:        8.0,
		feePerKg:        5.0,
		maxDistance:     100.0,
	}
}

func (s *InternalService) Enum() shipping.ShippingService {
	return s.shippingService
}

func (s *InternalService) GetCapabilities() (*shipping.ServiceCapabilities, error) {
	return &shipping.ServiceCapabilities{
		Provider:          s.shippingService.String(),
		APIVersion:        "v1.0",
		SupportedServices: []string{"standard", "express"},
		Coverage:          []string{"Metro Manila", "Cavite", "Laguna", "Rizal", "Bulacan"},
		Features: shipping.Features{
			RealTimeTracking:    false,
			RouteOptimization:   false,
			ScheduledDelivery:   true,
			SpecialRequests:     false,
			MultipleStops:       false,
			WeightBasedPricing:  true,
			Insurance:           false,
			ProofOfDelivery:     false,
			CashOnDelivery:      true,
			ContactlessDelivery: true,
		},
		Metadata: map[string]any{
			"base_fee":     s.baseFee,
			"fee_per_km":   s.feePerKm,
			"fee_per_kg":   s.feePerKg,
			"max_distance": s.maxDistance,
		},
	}, nil
}

func (s *InternalService) GetQuotation(req shipping.ShippingRequest) (*shipping.ShippingQuotation, error) {
	if req.PickupLocation.Coordinates.Lat == "" || req.PickupLocation.Coordinates.Lng == "" {
		return nil, errors.Join(errs.ErrShippingInvalidCoordinates, errors.New("pickup location"))
	}
	if req.DeliveryLocation.Coordinates.Lat == "" || req.DeliveryLocation.Coordinates.Lng == "" {
		return nil, errors.Join(errs.ErrShippingInvalidCoordinates, errors.New("delivery location"))
	}

	distance, err := s.calculateDistance(
		req.PickupLocation.Coordinates,
		req.DeliveryLocation.Coordinates,
	)
	if err != nil {
		return nil, errors.Join(errs.ErrShippingDistanceCalculation, err)
	}

	if distance > s.maxDistance {
		return nil, errors.Join(errs.ErrShippingDistanceExceeded, fmt.Errorf("%.2f km > %.2f km max", distance, s.maxDistance))
	}

	weight, err := s.parseWeight(req.Package.Weight)
	if err != nil {
		return nil, errors.Join(errs.ErrShippingInvalidWeight, err)
	}

	fee := s.calculateFee(distance, weight, req.ServiceType)

	serviceType := req.ServiceType
	if serviceType == "" {
		serviceType = "standard"
	}

	eta := s.calculateETA(distance, serviceType)

	return &shipping.ShippingQuotation{
		ID:           s.generateQuotationID(),
		Currency:     "PHP",
		ServiceType:  serviceType,
		Fee:          fee,
		DistanceKm:   distance,
		EstimatedETA: eta,
		ExpiresAt:    time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		Metadata: map[string]any{
			"base_fee":     s.baseFee,
			"distance_fee": distance * s.feePerKm,
			"weight_fee":   weight * s.feePerKg,
			"weight_kg":    weight,
		},
	}, nil
}

func (s *InternalService) CreateOrder(req shipping.ShippingRequest) (*shipping.ShippingOrder, error) {
	return nil, errs.ErrShippingNotImplemented
}

func (s *InternalService) GetOrderStatus(orderID string) (*shipping.ShippingOrder, error) {
	return nil, errs.ErrShippingNotImplemented
}

func (s *InternalService) CancelOrder(orderID string) error {
	return errs.ErrShippingNotImplemented
}

func (s *InternalService) calculateDistance(pickup, delivery shipping.Coordinates) (float64, error) {
	pickupLat, err := strconv.ParseFloat(pickup.Lat, 64)
	if err != nil {
		return 0, errors.Join(errs.ErrShippingInvalidLatitude, errors.New("pickup"))
	}
	pickupLng, err := strconv.ParseFloat(pickup.Lng, 64)
	if err != nil {
		return 0, errors.Join(errs.ErrShippingInvalidLongitude, errors.New("pickup"))
	}
	deliveryLat, err := strconv.ParseFloat(delivery.Lat, 64)
	if err != nil {
		return 0, errors.Join(errs.ErrShippingInvalidLatitude, errors.New("delivery"))
	}
	deliveryLng, err := strconv.ParseFloat(delivery.Lng, 64)
	if err != nil {
		return 0, errors.Join(errs.ErrShippingInvalidLongitude, errors.New("delivery"))
	}

	return haversineDistance(pickupLat, pickupLng, deliveryLat, deliveryLng), nil
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func (s *InternalService) parseWeight(weightStr string) (float64, error) {
	if weightStr == "" {
		return 1.0, nil
	}

	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		return 0, errors.Join(errs.ErrShippingInvalidWeight, err)
	}

	if weight <= 0 {
		return 0, errs.ErrShippingInvalidWeightRange
	}

	return weight, nil
}

func (s *InternalService) calculateFee(distance, weight float64, serviceType string) float64 {
	baseFee := s.baseFee
	distanceFee := distance * s.feePerKm
	weightFee := weight * s.feePerKg

	totalFee := baseFee + distanceFee + weightFee

	switch serviceType {
	case "express":
		totalFee *= 1.5
	case "standard":
	default:
	}

	return math.Round(totalFee*100) / 100
}

func (s *InternalService) calculateETA(distance float64, serviceType string) int {
	baseTimeMinutes := int((distance/30.0)*60 + 15)

	switch serviceType {
	case "express":
		return int(float64(baseTimeMinutes) * 0.7)
	case "standard":
		return baseTimeMinutes
	default:
		return baseTimeMinutes
	}
}

func (s *InternalService) generateQuotationID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("INT-%d", timestamp)
}

var _ shipping.IShippingService = (*InternalService)(nil)

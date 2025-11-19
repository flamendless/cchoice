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

type CChoiceService struct {
	shippingService  shipping.ShippingService
	baseFee          float64
	feePerKm         float64
	feePerKg         float64
	maxDistance      float64
	businessLocation *shipping.Location
}

func MustInit() *CChoiceService {
	cfg := conf.Conf()
	if cfg.ShippingService != shipping.SHIPPING_SERVICE_CCHOICE.String() {
		panic(errs.ErrCChoiceServiceInit)
	}

	return &CChoiceService{
		shippingService: shipping.SHIPPING_SERVICE_CCHOICE,
		baseFee:         50.0,
		feePerKm:        8.0,
		feePerKg:        5.0,
		maxDistance:     100.0,
		businessLocation: &shipping.Location{
			Coordinates: shipping.Coordinates{
				Lat: cfg.Business.Lat,
				Lng: cfg.Business.Lng,
			},
			Address: cfg.Business.Address,
		},
	}
}

func (s *CChoiceService) Enum() shipping.ShippingService {
	return s.shippingService
}

func (s *CChoiceService) GetBusinessLocation() *shipping.Location {
	return s.businessLocation
}

func (s *CChoiceService) GetCapabilities() (*shipping.ServiceCapabilities, error) {
	return &shipping.ServiceCapabilities{
		Provider:   s.shippingService.String(),
		APIVersion: "v1.0",
		SupportedServices: []shipping.ServiceType{
			shipping.SERVICE_TYPE_STANDARD,
			shipping.SERVICE_TYPE_EXPRESS,
			shipping.SERVICE_TYPE_2000KG_ALUMINUM,
			shipping.SERVICE_TYPE_2000KG_ALUMINUM_LD,
			shipping.SERVICE_TYPE_2000KG_FB,
			shipping.SERVICE_TYPE_2000KG_FB_LD,
			shipping.SERVICE_TYPE_2000KG_OPENTRUCK,
			shipping.SERVICE_TYPE_2000KG_OPENTRUCK_LD,
			shipping.SERVICE_TYPE_600KG_MPV,
			shipping.SERVICE_TYPE_600KG_MPV_LD,
			shipping.SERVICE_TYPE_MOTORCYCLE,
			shipping.SERVICE_TYPE_MPV,
			shipping.SERVICE_TYPE_MPV_INTERCITY,
			shipping.SERVICE_TYPE_SEDAN,
			shipping.SERVICE_TYPE_SEDAN_INTERCITY,
			shipping.SERVICE_TYPE_TRUCK330,
			shipping.SERVICE_TYPE_10WHEEL_TRUCK,
			shipping.SERVICE_TYPE_LD_10WHEEL_TRUCK,
			shipping.SERVICE_TYPE_TRUCK550,
			shipping.SERVICE_TYPE_VAN,
			shipping.SERVICE_TYPE_VAN1000,
			shipping.SERVICE_TYPE_VAN_INTERCITY,
			shipping.SERVICE_TYPE_PICKUP_800KG_INTERCITY,
		},
		Coverage: []string{"Metro Manila", "Cavite", "Laguna", "Rizal", "Bulacan"},
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

func (s *CChoiceService) GetQuotation(req shipping.ShippingRequest) (*shipping.ShippingQuotation, error) {
	if req.PickupLocation.Coordinates.Lat == "" || req.PickupLocation.Coordinates.Lng == "" {
		return nil, errors.Join(
			errs.ErrCChoice,
			errs.ErrShippingInvalidCoordinates,
			errs.ErrShippingPickupLocation,
		)
	}
	if req.DeliveryLocation.Coordinates.Lat == "" || req.DeliveryLocation.Coordinates.Lng == "" {
		return nil, errors.Join(
			errs.ErrCChoice,
			errs.ErrShippingInvalidCoordinates,
			errs.ErrShippingDeliveryLocation,
		)
	}

	distance, err := s.calculateDistance(
		req.PickupLocation.Coordinates,
		req.DeliveryLocation.Coordinates,
	)
	if err != nil {
		return nil, errors.Join(
			errs.ErrCChoice,
			errs.ErrShippingDistanceCalculation,
			err,
		)
	}

	if distance > s.maxDistance {
		return nil, errors.Join(
			errs.ErrCChoice,
			errs.ErrShippingDistanceExceeded,
			fmt.Errorf("%.2f km > %.2f km max", distance, s.maxDistance),
		)
	}

	weight, err := s.parseWeight(req.Package.Weight)
	if err != nil {
		return nil, errors.Join(
			errs.ErrCChoice,
			errs.ErrShippingInvalidWeight,
			err,
		)
	}

	fee := s.calculateFee(distance, weight, req.ServiceType)

	serviceType := req.ServiceType
	if serviceType == shipping.SERVICE_TYPE_UNDEFINED {
		serviceType = shipping.SERVICE_TYPE_STANDARD
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

func (s *CChoiceService) CreateOrder(req shipping.ShippingRequest) (*shipping.ShippingOrder, error) {
	return nil, errors.Join(errs.ErrCChoice, errs.ErrShippingNotImplemented)
}

func (s *CChoiceService) GetOrderStatus(orderID string) (*shipping.ShippingOrder, error) {
	return nil, errors.Join(errs.ErrCChoice, errs.ErrShippingNotImplemented)
}

func (s *CChoiceService) CancelOrder(orderID string) error {
	return errors.Join(errs.ErrCChoice, errs.ErrShippingNotImplemented)
}

func (s *CChoiceService) calculateDistance(pickup, delivery shipping.Coordinates) (float64, error) {
	pickupLat, err := strconv.ParseFloat(pickup.Lat, 64)
	if err != nil {
		return 0, errors.Join(
			errs.ErrCChoice,
			errs.ErrShippingInvalidLatitude,
			errs.ErrShippingPickupLocation,
			err,
		)
	}
	pickupLng, err := strconv.ParseFloat(pickup.Lng, 64)
	if err != nil {
		return 0, errors.Join(
			errs.ErrCChoice,
			errs.ErrShippingInvalidLongitude,
			errs.ErrShippingPickupLocation,
			err,
		)
	}
	deliveryLat, err := strconv.ParseFloat(delivery.Lat, 64)
	if err != nil {
		return 0, errors.Join(
			errs.ErrCChoice,
			errs.ErrShippingInvalidLatitude,
			errs.ErrShippingDeliveryLocation,
			err,
		)
	}
	deliveryLng, err := strconv.ParseFloat(delivery.Lng, 64)
	if err != nil {
		return 0, errors.Join(
			errs.ErrCChoice,
			errs.ErrShippingInvalidLongitude,
			errs.ErrShippingDeliveryLocation,
			err,
		)
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

func (s *CChoiceService) parseWeight(weightStr string) (float64, error) {
	if weightStr == "" {
		return 1.0, nil
	}

	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		return 0, errors.Join(
			errs.ErrCChoice,
			errs.ErrShippingInvalidWeight,
			err,
		)
	}

	if weight <= 0 {
		return 0, errors.Join(errs.ErrCChoice, errs.ErrShippingInvalidWeightRange)
	}

	return weight, nil
}

func (s *CChoiceService) calculateFee(distance, weight float64, serviceType shipping.ServiceType) float64 {
	baseFee := s.baseFee
	distanceFee := distance * s.feePerKm
	weightFee := weight * s.feePerKg

	totalFee := baseFee + distanceFee + weightFee

	switch serviceType {
	case shipping.SERVICE_TYPE_EXPRESS:
		totalFee *= 1.5
	case shipping.SERVICE_TYPE_MOTORCYCLE:
		totalFee *= 0.8 // Motorcycle is typically cheaper
	case shipping.SERVICE_TYPE_SEDAN, shipping.SERVICE_TYPE_SEDAN_INTERCITY:
		totalFee *= 1.1 // Car services
	case shipping.SERVICE_TYPE_MPV, shipping.SERVICE_TYPE_MPV_INTERCITY, shipping.SERVICE_TYPE_600KG_MPV, shipping.SERVICE_TYPE_600KG_MPV_LD:
		totalFee *= 1.2 // MPV services
	case shipping.SERVICE_TYPE_VAN, shipping.SERVICE_TYPE_VAN1000, shipping.SERVICE_TYPE_VAN_INTERCITY:
		totalFee *= 1.4 // Van services
	case shipping.SERVICE_TYPE_TRUCK330, shipping.SERVICE_TYPE_TRUCK550, shipping.SERVICE_TYPE_10WHEEL_TRUCK, shipping.SERVICE_TYPE_LD_10WHEEL_TRUCK:
		totalFee *= 1.8 // Truck services
	case shipping.SERVICE_TYPE_PICKUP_800KG_INTERCITY:
		totalFee *= 1.3 // Pickup services
	case shipping.SERVICE_TYPE_2000KG_ALUMINUM, shipping.SERVICE_TYPE_2000KG_ALUMINUM_LD, shipping.SERVICE_TYPE_2000KG_FB, shipping.SERVICE_TYPE_2000KG_FB_LD, shipping.SERVICE_TYPE_2000KG_OPENTRUCK, shipping.SERVICE_TYPE_2000KG_OPENTRUCK_LD:
		totalFee *= 2.2 // Heavy duty services
	case shipping.SERVICE_TYPE_STANDARD:
	default:
	}

	return math.Round(totalFee*100) / 100
}

func (s *CChoiceService) calculateETA(distance float64, serviceType shipping.ServiceType) int {
	baseTimeMinutes := int((distance/30.0)*60 + 15)

	switch serviceType {
	case shipping.SERVICE_TYPE_EXPRESS:
		return int(float64(baseTimeMinutes) * 0.7)
	case shipping.SERVICE_TYPE_MOTORCYCLE:
		return int(float64(baseTimeMinutes) * 0.6) // Motorcycle is typically fastest
	case shipping.SERVICE_TYPE_SEDAN, shipping.SERVICE_TYPE_SEDAN_INTERCITY:
		return int(float64(baseTimeMinutes) * 0.8) // Car services - fast but limited capacity
	case shipping.SERVICE_TYPE_MPV, shipping.SERVICE_TYPE_MPV_INTERCITY, shipping.SERVICE_TYPE_600KG_MPV, shipping.SERVICE_TYPE_600KG_MPV_LD:
		return int(float64(baseTimeMinutes) * 0.9) // MPV services - moderate speed
	case shipping.SERVICE_TYPE_VAN, shipping.SERVICE_TYPE_VAN1000, shipping.SERVICE_TYPE_VAN_INTERCITY:
		return int(float64(baseTimeMinutes) * 1.1) // Van services - slower due to size
	case shipping.SERVICE_TYPE_TRUCK330, shipping.SERVICE_TYPE_TRUCK550, shipping.SERVICE_TYPE_10WHEEL_TRUCK, shipping.SERVICE_TYPE_LD_10WHEEL_TRUCK:
		return int(float64(baseTimeMinutes) * 1.3) // Truck services - slower heavy vehicle
	case shipping.SERVICE_TYPE_PICKUP_800KG_INTERCITY:
		return baseTimeMinutes // Pickup services - same as standard
	case shipping.SERVICE_TYPE_2000KG_ALUMINUM, shipping.SERVICE_TYPE_2000KG_ALUMINUM_LD, shipping.SERVICE_TYPE_2000KG_FB, shipping.SERVICE_TYPE_2000KG_FB_LD, shipping.SERVICE_TYPE_2000KG_OPENTRUCK, shipping.SERVICE_TYPE_2000KG_OPENTRUCK_LD:
		return int(float64(baseTimeMinutes) * 1.5) // Heavy duty services - slowest due to size and weight
	case shipping.SERVICE_TYPE_STANDARD:
		return baseTimeMinutes
	default:
		return baseTimeMinutes
	}
}

func (s *CChoiceService) generateQuotationID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("INT-%d", timestamp)
}

var _ shipping.IShippingService = (*CChoiceService)(nil)

package errs

import "errors"

var (
	ErrShippingNotImplemented      = errors.New("[SHIPPING]: Operation not implemented")
	ErrShippingInvalidCoordinates  = errors.New("[SHIPPING]: Invalid coordinates provided")
	ErrShippingInvalidWeight       = errors.New("[SHIPPING]: Invalid package weight")
	ErrShippingDistanceExceeded    = errors.New("[SHIPPING]: Delivery distance exceeds service area")
	ErrShippingServiceInit         = errors.New("[SHIPPING]: Failed to initialize shipping service")
	ErrShippingInvalidLatitude     = errors.New("[SHIPPING]: Invalid latitude value")
	ErrShippingInvalidLongitude    = errors.New("[SHIPPING]: Invalid longitude value")
	ErrShippingInvalidWeightRange  = errors.New("[SHIPPING]: Weight must be greater than zero")
	ErrShippingDistanceCalculation = errors.New("[SHIPPING]: Failed to calculate distance")
	ErrShippingPickupLocation      = errors.New("[SHIPPING]: Pickup location")
	ErrShippingDeliveryLocation    = errors.New("[SHIPPING]: Delivery location")
)

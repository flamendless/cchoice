package shipping

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=ShippingService -trimprefix=SHIPPING_SERVICE_

type ShippingService int

const (
	SHIPPING_SERVICE_UNDEFINED ShippingService = iota
	SHIPPING_SERVICE_LALAMOVE
)

func ParseShippingServiceToEnum(ss string) ShippingService {
	switch strings.ToUpper(ss) {
	case SHIPPING_SERVICE_LALAMOVE.String():
		return SHIPPING_SERVICE_LALAMOVE
	default:
		panic(fmt.Errorf("undefined shipping service '%s'", ss))
	}
}

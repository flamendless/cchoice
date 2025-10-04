//go:generate go tool stringer -type=ShippingService -trimprefix=SHIPPING_SERVICE_

package shipping

import (
	"fmt"
	"strings"
)

type ShippingService int

const (
	SHIPPING_SERVICE_UNDEFINED ShippingService = iota
	SHIPPING_SERVICE_LALAMOVE
	SHIPPING_SERVICE_CCHOICE
)

func ParseShippingServiceToEnum(ss string) ShippingService {
	switch strings.ToUpper(ss) {
	case SHIPPING_SERVICE_CCHOICE.String():
		return SHIPPING_SERVICE_CCHOICE
	case SHIPPING_SERVICE_LALAMOVE.String():
		return SHIPPING_SERVICE_LALAMOVE
	default:
		panic(fmt.Errorf("undefined shipping service '%s'", ss))
	}
}

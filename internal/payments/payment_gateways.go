package payments

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=PaymentGateway -trimprefix=PAYMENT_GATEWAY_

type PaymentGateway int

const (
	PAYMENT_GATEWAY_UNDEFINED PaymentGateway = iota
	PAYMENT_GATEWAY_PAYMONGO
)

func ParsePaymentGatewayToEnum(pg string) PaymentGateway {
	switch strings.ToUpper(pg) {
	case PAYMENT_GATEWAY_PAYMONGO.String():
		return PAYMENT_GATEWAY_PAYMONGO
	default:
		panic(fmt.Errorf("undefined payment gateway '%s'", pg))
	}
}

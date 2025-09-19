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

func (pg PaymentGateway) Code() string {
	switch pg {
	case PAYMENT_GATEWAY_PAYMONGO:
		return "pm"
	default:
		return "px"
	}
}

func (pg PaymentGateway) GetAllPaymentMethods() []PaymentMethod {
	return []PaymentMethod{
		PAYMENT_METHOD_QRPH,
		PAYMENT_METHOD_BILLEASE,
		PAYMENT_METHOD_CARD,
		PAYMENT_METHOD_DOB,
		PAYMENT_METHOD_DOB_UBP,
		PAYMENT_METHOD_BRANKAS_BDO,
		PAYMENT_METHOD_BRANKAS_LANDBANK,
		PAYMENT_METHOD_BRANKAS_METROBANK,
		PAYMENT_METHOD_GCASH,
		PAYMENT_METHOD_GRAB_PAY,
		PAYMENT_METHOD_PAYMAYA,
	}
}

func (pg PaymentGateway) GetPrioritizedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{
		PAYMENT_METHOD_QRPH,
		PAYMENT_METHOD_DOB_UBP,
		PAYMENT_METHOD_BRANKAS_BDO,
		PAYMENT_METHOD_GCASH,
		PAYMENT_METHOD_PAYMAYA,
	}
}

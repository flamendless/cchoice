package enums

import "strings"

//go:generate go tool stringer -type=PaymentTermsUnit -trimprefix=PAYMENT_TERMS_

type PaymentTermsUnit int

const (
	PAYMENT_TERMS_UNDEFINED PaymentTermsUnit = iota
	PAYMENT_TERMS_DAYS
	PAYMENT_TERMS_MONTHS
	PAYMENT_TERMS_YEARS
)

func ParsePaymentTermsUnitToEnum(e string) PaymentTermsUnit {
	switch strings.ToUpper(e) {
	case PAYMENT_TERMS_DAYS.String():
		return PAYMENT_TERMS_DAYS
	case PAYMENT_TERMS_MONTHS.String():
		return PAYMENT_TERMS_MONTHS
	case PAYMENT_TERMS_YEARS.String():
		return PAYMENT_TERMS_YEARS
	default:
		return PAYMENT_TERMS_UNDEFINED
	}
}

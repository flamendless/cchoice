package enums

import "strings"

//go:generate go tool stringer -type=ReceiptPaymentForm -trimprefix=RECEIPT_PAYMENT_

type ReceiptPaymentForm int

const (
	RECEIPT_PAYMENT_UNDEFINED ReceiptPaymentForm = iota
	RECEIPT_PAYMENT_CASH
	RECEIPT_PAYMENT_CHECK
)

func ParseReceiptPaymentFormToEnum(e string) ReceiptPaymentForm {
	switch strings.ToUpper(e) {
	case RECEIPT_PAYMENT_CASH.String():
		return RECEIPT_PAYMENT_CASH
	case RECEIPT_PAYMENT_CHECK.String():
		return RECEIPT_PAYMENT_CHECK
	default:
		return RECEIPT_PAYMENT_UNDEFINED
	}
}

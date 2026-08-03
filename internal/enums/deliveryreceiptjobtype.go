package enums

import "strings"

//go:generate go tool stringer -type=DeliveryReceiptJobType -trimprefix=DELIVERY_RECEIPT_JOB_

type DeliveryReceiptJobType int

const (
	DELIVERY_RECEIPT_JOB_UNDEFINED DeliveryReceiptJobType = iota
	DELIVERY_RECEIPT_JOB_GENERATE_PDF
	DELIVERY_RECEIPT_JOB_SEND_EMAIL
)

func ParseDeliveryReceiptJobTypeToEnum(e string) DeliveryReceiptJobType {
	switch strings.ToUpper(e) {
	case DELIVERY_RECEIPT_JOB_GENERATE_PDF.String():
		return DELIVERY_RECEIPT_JOB_GENERATE_PDF
	case DELIVERY_RECEIPT_JOB_SEND_EMAIL.String():
		return DELIVERY_RECEIPT_JOB_SEND_EMAIL
	default:
		return DELIVERY_RECEIPT_JOB_UNDEFINED
	}
}

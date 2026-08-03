package enums

import "strings"

//go:generate go tool stringer -type=ReceiptStatus -trimprefix=RECEIPT_STATUS_

type ReceiptStatus int

const (
	RECEIPT_STATUS_UNDEFINED ReceiptStatus = iota
	RECEIPT_STATUS_DRAFT
	RECEIPT_STATUS_PROCESSING
	RECEIPT_STATUS_ISSUED
	RECEIPT_STATUS_SENT
	RECEIPT_STATUS_CANCELLED
)

func ParseReceiptStatusToEnum(e string) ReceiptStatus {
	switch strings.ToUpper(e) {
	case RECEIPT_STATUS_DRAFT.String():
		return RECEIPT_STATUS_DRAFT
	case RECEIPT_STATUS_PROCESSING.String():
		return RECEIPT_STATUS_PROCESSING
	case RECEIPT_STATUS_ISSUED.String():
		return RECEIPT_STATUS_ISSUED
	case RECEIPT_STATUS_SENT.String():
		return RECEIPT_STATUS_SENT
	case RECEIPT_STATUS_CANCELLED.String():
		return RECEIPT_STATUS_CANCELLED
	default:
		return RECEIPT_STATUS_UNDEFINED
	}
}

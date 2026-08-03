package enums

import "strings"

//go:generate go tool stringer -type=InvoiceStatus -trimprefix=INVOICE_STATUS_

type InvoiceStatus int

const (
	INVOICE_STATUS_UNDEFINED InvoiceStatus = iota
	INVOICE_STATUS_DRAFT
	INVOICE_STATUS_PROCESSING
	INVOICE_STATUS_ISSUED
	INVOICE_STATUS_SENT
	INVOICE_STATUS_PAID
	INVOICE_STATUS_CANCELLED
)

func ParseInvoiceStatusToEnum(e string) InvoiceStatus {
	switch strings.ToUpper(e) {
	case INVOICE_STATUS_DRAFT.String():
		return INVOICE_STATUS_DRAFT
	case INVOICE_STATUS_PROCESSING.String():
		return INVOICE_STATUS_PROCESSING
	case INVOICE_STATUS_ISSUED.String():
		return INVOICE_STATUS_ISSUED
	case INVOICE_STATUS_SENT.String():
		return INVOICE_STATUS_SENT
	case INVOICE_STATUS_PAID.String():
		return INVOICE_STATUS_PAID
	case INVOICE_STATUS_CANCELLED.String():
		return INVOICE_STATUS_CANCELLED
	default:
		return INVOICE_STATUS_UNDEFINED
	}
}

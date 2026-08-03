package enums

import "strings"

//go:generate go tool stringer -type=InvoiceJobType -trimprefix=INVOICE_JOB_

type InvoiceJobType int

const (
	INVOICE_JOB_UNDEFINED InvoiceJobType = iota
	INVOICE_JOB_GENERATE_PDF
	INVOICE_JOB_SEND_EMAIL
)

func ParseInvoiceJobTypeToEnum(e string) InvoiceJobType {
	switch strings.ToUpper(e) {
	case INVOICE_JOB_GENERATE_PDF.String():
		return INVOICE_JOB_GENERATE_PDF
	case INVOICE_JOB_SEND_EMAIL.String():
		return INVOICE_JOB_SEND_EMAIL
	default:
		return INVOICE_JOB_UNDEFINED
	}
}

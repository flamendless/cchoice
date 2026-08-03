package enums

import "strings"

//go:generate go tool stringer -type=InvoiceJobStatus -trimprefix=INVOICE_JOB_STATUS_

type InvoiceJobStatus int

const (
	INVOICE_JOB_STATUS_UNDEFINED InvoiceJobStatus = iota
	INVOICE_JOB_STATUS_PENDING
	INVOICE_JOB_STATUS_PROCESSING
	INVOICE_JOB_STATUS_COMPLETED
	INVOICE_JOB_STATUS_FAILED
)

func ParseInvoiceJobStatusToEnum(e string) InvoiceJobStatus {
	switch strings.ToUpper(e) {
	case INVOICE_JOB_STATUS_PENDING.String():
		return INVOICE_JOB_STATUS_PENDING
	case INVOICE_JOB_STATUS_PROCESSING.String():
		return INVOICE_JOB_STATUS_PROCESSING
	case INVOICE_JOB_STATUS_COMPLETED.String():
		return INVOICE_JOB_STATUS_COMPLETED
	case INVOICE_JOB_STATUS_FAILED.String():
		return INVOICE_JOB_STATUS_FAILED
	default:
		return INVOICE_JOB_STATUS_UNDEFINED
	}
}

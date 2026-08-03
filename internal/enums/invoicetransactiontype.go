package enums

import "strings"

//go:generate go tool stringer -type=InvoiceTransactionType -trimprefix=INVOICE_TRANSACTION_

type InvoiceTransactionType int

const (
	INVOICE_TRANSACTION_UNDEFINED InvoiceTransactionType = iota
	INVOICE_TRANSACTION_CASH_SALES
	INVOICE_TRANSACTION_CHARGE_SALES
)

func ParseInvoiceTransactionTypeToEnum(e string) InvoiceTransactionType {
	switch strings.ToUpper(e) {
	case INVOICE_TRANSACTION_CASH_SALES.String():
		return INVOICE_TRANSACTION_CASH_SALES
	case INVOICE_TRANSACTION_CHARGE_SALES.String():
		return INVOICE_TRANSACTION_CHARGE_SALES
	default:
		return INVOICE_TRANSACTION_UNDEFINED
	}
}

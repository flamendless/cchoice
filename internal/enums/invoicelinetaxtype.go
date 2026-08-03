package enums

import "strings"

//go:generate go tool stringer -type=InvoiceLineTaxType -trimprefix=INVOICE_LINE_TAX_

type InvoiceLineTaxType int

const (
	INVOICE_LINE_TAX_UNDEFINED InvoiceLineTaxType = iota
	INVOICE_LINE_TAX_VATABLE
	INVOICE_LINE_TAX_VAT_EXEMPT
	INVOICE_LINE_TAX_ZERO_RATED
)

func ParseInvoiceLineTaxTypeToEnum(e string) InvoiceLineTaxType {
	switch strings.ToUpper(e) {
	case INVOICE_LINE_TAX_VATABLE.String():
		return INVOICE_LINE_TAX_VATABLE
	case INVOICE_LINE_TAX_VAT_EXEMPT.String():
		return INVOICE_LINE_TAX_VAT_EXEMPT
	case INVOICE_LINE_TAX_ZERO_RATED.String():
		return INVOICE_LINE_TAX_ZERO_RATED
	default:
		return INVOICE_LINE_TAX_UNDEFINED
	}
}

func (t InvoiceLineTaxType) Label() string {
	switch t {
	case INVOICE_LINE_TAX_VAT_EXEMPT:
		return "Exempt"
	case INVOICE_LINE_TAX_ZERO_RATED:
		return "Zero"
	default:
		return "VAT"
	}
}

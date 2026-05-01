package enums

import "strings"

//go:generate go tool stringer -type=QuotationStatus -trimprefix=QUOTATION_STATUS_

type QuotationStatus int

const (
	QUOTATION_STATUS_UNDEFINED QuotationStatus = iota
	QUOTATION_STATUS_DRAFT
	QUOTATION_STATUS_IN_REVIEW
	QUOTATION_STATUS_APPROVED
	QUOTATION_STATUS_COMPLETED
)

func ParseQuotationStatus(s string) QuotationStatus {
	switch strings.ToUpper(s) {
	case QUOTATION_STATUS_DRAFT.String():
		return QUOTATION_STATUS_DRAFT
	case QUOTATION_STATUS_IN_REVIEW.String():
		return QUOTATION_STATUS_IN_REVIEW
	case QUOTATION_STATUS_APPROVED.String():
		return QUOTATION_STATUS_APPROVED
	case QUOTATION_STATUS_COMPLETED.String():
		return QUOTATION_STATUS_COMPLETED
	default:
		return QUOTATION_STATUS_UNDEFINED
	}
}

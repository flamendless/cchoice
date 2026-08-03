package enums

import "strings"

//go:generate go tool stringer -type=CollectionReceiptJobType -trimprefix=COLLECTION_RECEIPT_JOB_

type CollectionReceiptJobType int

const (
	COLLECTION_RECEIPT_JOB_UNDEFINED CollectionReceiptJobType = iota
	COLLECTION_RECEIPT_JOB_GENERATE_PDF
	COLLECTION_RECEIPT_JOB_SEND_EMAIL
)

func ParseCollectionReceiptJobTypeToEnum(e string) CollectionReceiptJobType {
	switch strings.ToUpper(e) {
	case COLLECTION_RECEIPT_JOB_GENERATE_PDF.String():
		return COLLECTION_RECEIPT_JOB_GENERATE_PDF
	case COLLECTION_RECEIPT_JOB_SEND_EMAIL.String():
		return COLLECTION_RECEIPT_JOB_SEND_EMAIL
	default:
		return COLLECTION_RECEIPT_JOB_UNDEFINED
	}
}

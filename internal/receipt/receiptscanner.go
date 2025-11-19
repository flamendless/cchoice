package receipt

import (
	"cchoice/internal/errs"
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=ReceiptScanner -trimprefix=RECEIPT_SCANNER_

type ReceiptScanner int

const (
	RECEIPT_SCANNER_UNDEFINED ReceiptScanner = iota
	RECEIPT_SCANNER_GOOGLEVISION
)

func ParseStorageProviderToEnum(sp string) ReceiptScanner {
	switch strings.ToUpper(sp) {
	case RECEIPT_SCANNER_GOOGLEVISION.String():
		return RECEIPT_SCANNER_GOOGLEVISION
	default:
		panic(fmt.Errorf("%w: '%s'", errs.ErrCmdUndefinedService, sp))
	}
}

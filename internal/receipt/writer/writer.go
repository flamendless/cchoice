package writer

import (
	"cchoice/internal/errs"
	"cchoice/internal/receipt/scanner"
	"fmt"
	"strings"
)

type OutputFormat int

const (
	OutputFormatPrint OutputFormat = iota
	OutputFormatCSV
	OutputFormatJSON
)

func (f OutputFormat) String() string {
	switch f {
	case OutputFormatPrint:
		return "print"
	case OutputFormatCSV:
		return "csv"
	case OutputFormatJSON:
		return "json"
	default:
		return "unknown"
	}
}

func ParseOutputFormat(s string) (OutputFormat, error) {
	switch strings.ToLower(s) {
	case "print":
		return OutputFormatPrint, nil
	case "csv":
		return OutputFormatCSV, nil
	case "json":
		return OutputFormatJSON, nil
	default:
		return OutputFormatPrint, fmt.Errorf("%w: %s", errs.ErrReceiptInvalidFormat, s)
	}
}

type IReceiptWriter interface {
	Write(data *scanner.ReceiptData, outputPath string) error
}

func valueOrEmpty(value string) string {
	if value == "" {
		return "(not found)"
	}
	return value
}

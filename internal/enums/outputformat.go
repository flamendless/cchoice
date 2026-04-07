package enums

import (
	"strings"
)

//go:generate go tool stringer -type=OutputFormat -trimprefix=OUTPUT_FORMAT_

type OutputFormat int

const (
	OUTPUT_FORMAT_UNDEFINED OutputFormat = iota
	OUTPUT_FORMAT_CSV
	OUTPUT_FORMAT_XLSX
)

func ParseOutputFormatToEnum(format string) OutputFormat {
	switch strings.ToUpper(format) {
	case OUTPUT_FORMAT_CSV.String():
		return OUTPUT_FORMAT_CSV
	case OUTPUT_FORMAT_XLSX.String():
		return OUTPUT_FORMAT_XLSX
	default:
		return OUTPUT_FORMAT_UNDEFINED
	}
}

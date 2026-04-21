package enums

import "strings"

//go:generate go tool stringer -type=StocksIn -trimprefix=STOCKS_IN_

type StocksIn int

const (
	STOCKS_IN_UNDEFINED StocksIn = iota
	STOCKS_IN_OFFICE
	STOCKS_IN_SUPPLIER
)

var AllStocksIn = []StocksIn{
	STOCKS_IN_OFFICE,
	STOCKS_IN_SUPPLIER,
}

func ParseStocksInToEnum(e string) StocksIn {
	switch e {
	case STOCKS_IN_OFFICE.String():
		return STOCKS_IN_OFFICE
	case STOCKS_IN_SUPPLIER.String():
		return STOCKS_IN_SUPPLIER
	default:
		return STOCKS_IN_UNDEFINED
	}
}

func MustParseStocksInToEnum(e string) StocksIn {
	switch strings.ToUpper(e) {
	case STOCKS_IN_OFFICE.String():
		return STOCKS_IN_OFFICE
	case STOCKS_IN_SUPPLIER.String():
		return STOCKS_IN_SUPPLIER
	default:
		panic("Invalid StocksIn. Got '" + e + "'")
	}
}

func (s StocksIn) IsValid() bool {
	return s != STOCKS_IN_UNDEFINED
}

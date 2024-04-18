package utils

import (
	"cchoice/internal/domains/parser"
	"strings"

	"github.com/Rhymond/go-money"
	"github.com/shopspring/decimal"
)

func SanitizePrice(price string) (*money.Money, []error) {
	var errs []error = make([]error, 0, 2)
	currency := money.PHP

	hasPHPTrailing := strings.HasSuffix(price, "PHP")
	if hasPHPTrailing {
		price = strings.TrimSuffix(price, "PHP")
	}

	hasPHPLeading := strings.HasPrefix(price, "PHP")
	if hasPHPLeading {
		price = strings.TrimPrefix(price, "PHP")
	}

	price = strings.TrimSpace(price)
	price = strings.Replace(price, ",", "", 1)

	errPrice := ValidateNotBlank(price, "unit price")
	if errPrice != nil {
		errs = append(errs, errPrice)
	}

	unitPrice, err := decimal.NewFromString(price)
	if err != nil {
		parserErr := parser.NewParserError(parser.CantCovert, err.Error())
		errs = append(errs, parserErr)
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return money.New(unitPrice.CoefficientInt64(), currency), nil
}

func SanitizeColours(color string) string {
	if color == "-" {
		color = ""
	}
	return color
}

func SanitizeSize(size string) string {
	if size == "-" {
		size = ""
	}
	return size
}

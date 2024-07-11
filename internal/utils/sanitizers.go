package utils

import (
	"cchoice/internal/errs"
	"strings"

	"github.com/Rhymond/go-money"
	"github.com/shopspring/decimal"
)

func SanitizePrice(price string) (*money.Money, []error) {
	var errsRes []error = make([]error, 0, 2)
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
		errsRes = append(errsRes, errPrice)
	}

	unitPrice, err := decimal.NewFromString(price)
	if err != nil {
		parserErr := errs.NewParserError(errs.CantCovert, err.Error())
		errsRes = append(errsRes, parserErr)
	}

	if len(errsRes) > 0 {
		return nil, errsRes
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

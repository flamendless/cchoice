package utils

import (
	"errors"
	"strings"

	"github.com/Rhymond/go-money"
	"github.com/shopspring/decimal"
)

func SanitizePrice(price string) (*money.Money, error) {
	var errs error
	currency := money.PHP

	hasPHP := strings.HasSuffix(price, "PHP")
	if hasPHP {
		price = strings.TrimSuffix(price, "PHP")
		currency = money.PHP
	}

	price = strings.TrimSpace(price)

	errPrice := ValidateNotBlank(price, "unit price")
	if errPrice != nil {
		errs = errors.Join(errs, errPrice)
	}

	unitPrice, err := decimal.NewFromString(price)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return nil, errs
	}

	return money.NewFromFloat(unitPrice.InexactFloat64(), currency), nil
}

func SanitizeSize(size string) string {
	if size == "-" {
		size = ""
	}
	return size
}

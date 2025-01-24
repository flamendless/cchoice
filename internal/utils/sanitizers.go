package utils

import (
	"strings"
	"sync"

	"github.com/Rhymond/go-money"
	"github.com/gertd/go-pluralize"
)

var pluralizerOnce sync.Once
var pluralizer *pluralize.Client

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

	m, err := NewMoneyFromString(price, currency)
	if err != nil {
		errsRes = append(errsRes, err)
	}

	if len(errsRes) > 0 {
		return nil, errsRes
	}

	return m, nil
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

func SanitizeCategory(input string) string {
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "  ", " ")
	input = strings.ReplaceAll(input, "\\", " ")
	input = strings.ReplaceAll(input, "/", " ")
	if strings.Contains(input, "(") {
		idxLeft := strings.Index(input, "(")
		input = input[0:idxLeft]
	}
	input = strings.TrimSuffix(input, "AND")
	input = strings.Trim(input, "  ")
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "- ", "-")
	input = strings.ReplaceAll(input, " - ", "-")
	input = strings.ReplaceAll(input, " -", "-")

	pluralizerOnce.Do(func() {
		pluralizer = pluralize.NewClient()
	})
	if !pluralizer.IsPlural(input) {
		input = pluralizer.Plural(input)
	}

	return input
}

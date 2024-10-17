package utils

import (
	"cchoice/internal/constants"
	"cchoice/internal/errs"

	"github.com/Rhymond/go-money"
	"github.com/govalues/decimal"
)

func NewMoney(price int64, currency string) *money.Money {
	dec := decimal.MustNew(price, constants.DEC_SCALE)
	m := money.New(int64(dec.Coef()), currency)
	return m
}

func NewMoneyFromString(price string, currency string) (*money.Money, error) {
	unitPrice, err := decimal.Parse(price)
	if err != nil {
		parserErr := errs.NewParserError(errs.CantCovert, err.Error())
		return nil, parserErr
	}
	m := money.New(int64(unitPrice.Coef()), currency)
	return m, nil
}

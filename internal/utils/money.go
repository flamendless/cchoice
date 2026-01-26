package utils

import (
	"cchoice/internal/constants"
	"cchoice/internal/errs"
	"database/sql"
	"fmt"

	"github.com/Rhymond/go-money"
	"github.com/govalues/decimal"
)

func NewMoney(price int64, currency string) *money.Money {
	dec := decimal.MustNew(price, constants.DecimalScale)
	m := money.New(int64(dec.Coef()), currency)
	return m
}

func NewMoneyFromString(price string, currency string) (*money.Money, error) {
	unitPrice, err := decimal.Parse(price)
	if err != nil {
		parserErr := errs.NewParserError(errs.CantCovert, "%s", err.Error())
		return nil, parserErr
	}
	m := money.New(int64(unitPrice.Coef()), currency)
	return m, nil
}

//INFO: (Brandon) - 2nd return value 'price' can also be the same as the origPrice
func GetOrigAndDiscounted(
	isOnSale int64,
	unitPriceWithVat int64,
	unitPriceWithVatCurrency string,
	salePriceWithVat         sql.NullInt64,
	salePriceWithVatCurrency sql.NullString,
) (*money.Money, *money.Money, string) {
	origPrice := NewMoney(unitPriceWithVat, unitPriceWithVatCurrency)
	var price *money.Money
	var discountPercentage string
	if isOnSale == 1 {
		price = NewMoney(salePriceWithVat.Int64, salePriceWithVatCurrency.String)
		discount := ((unitPriceWithVat - salePriceWithVat.Int64) * 100.0) / unitPriceWithVat
		discountPercentage = fmt.Sprintf("%d%%", discount)
	} else {
		price = origPrice
	}
	return origPrice, price, discountPercentage
}

func GetDiscountAmount(
	isOnSale int64,
	unitPriceWithVat int64,
	unitPriceWithVatCurrency string,
	salePriceWithVat         sql.NullInt64,
	salePriceWithVatCurrency sql.NullString,
) *money.Money {
	if isOnSale != 1 {
		return NewMoney(0, "PHP")
	}

	discount := unitPriceWithVat - salePriceWithVat.Int64
	return NewMoney(discount, "PHP")
}

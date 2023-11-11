package model

import "github.com/shopspring/decimal"

type ConversionRate struct {
	BaseModel

	CurrencyID UUID
	Currency   Currency
	ToUSD      decimal.Decimal
	ToVND      decimal.Decimal
}

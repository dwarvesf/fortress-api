package model

import "time"

type Liability struct {
	BaseModel

	PaidAt     *time.Time `json:"paidAt"`
	Name       string     `json:"name"`
	Total      float64    `json:"total"`
	CurrencyID UUID       `json:"currencyID"`
}

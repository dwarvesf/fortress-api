package model

import "time"

type SalaryAdvance struct {
	BaseModel `json:"base_model"`

	EmployeeID UUID      `json:"employee_id"`
	Employee   *Employee `json:"employee"`

	CurrencyID UUID      `json:"currency_id"`
	Currency   *Currency `json:"currency"`

	AmountIcy      int64      `json:"amount_icy"`
	AmountUSD      float64    `json:"amount_usd"`
	BaseAmount     float64    `json:"base_amount"`
	ConversionRate float64    `json:"conversion_rate"`
	IsPaidBack     bool       `json:"is_paid_back"`
	PaidAt         *time.Time `json:"paidAt"`
}

func (SalaryAdvance) TableName() string { return "salary_advance_histories" }

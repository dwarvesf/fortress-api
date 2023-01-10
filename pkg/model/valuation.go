// please edit this file only with approval from hnh
package model

type Valuation struct {
	// basic info
	Year     string  `json:"year"`
	Rate     float64 `json:"rate"`
	Currency string  `json:"currency"`

	// valuation info
	Assets float64 `json:"assets"`

	// money that company will receive in the future
	AccountReceivable struct {
		Total float64          `json:"total"`
		Items []AccountingItem `json:"items"`
	} `json:"accountReceivable"`

	// money that company will pay in the future
	Liabilities struct {
		Total float64          `json:"total"`
		Items []AccountingItem `json:"items"`
	} `json:"liabilities"`

	// Total paid invoice, investment & bank interest
	Income struct {
		Total  float64 `json:"total"`
		Detail struct {
			ConsultantService float64 `json:"consultantService"`
			Investment        float64 `json:"investment"`
			Interest          float64 `json:"interest"`
		} `json:"detail"`
	} `json:"income"`

	// Sum of Expenses and payroll
	Outcome struct {
		Total  float64 `json:"total"`
		Detail struct {
			Payroll    float64 `json:"payroll"`
			Expense    float64 `json:"expense"`
			Investment float64 `json:"investment"`
		} `json:"detail"`
	} `json:"outcome"`
}

type AccountingItem struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type CurrencyView struct {
	USD float64
	VND float64
	EUR float64
	GBP float64
	SGD float64
}

package payroll

import "github.com/dwarvesf/fortress-api/pkg/model"

type payrollResponse struct {
	DisplayName          string            `json:"display_name"`
	BaseSalary           int64             `json:"base_salary"`
	SalaryAdvanceAmount  float64           `json:"salary_advance_amount"`
	Bonus                float64           `json:"bonus"`
	TotalWithoutContract float64           `json:"total_without_contract"`
	TotalWithContract    model.VietnamDong `json:"total_with_contract"`
	Notes                []string          `json:"notes"`
	Date                 int               `json:"date"`
	Month                int               `json:"month"`
	Year                 int               `json:"year"`
	BankAccountNumber    string            `json:"bank_account_number"`
	TWRecipientID        string            `json:"tw_recipient_id"` // will be removed
	TWRecipientName      string            `json:"tw_recipient_name"`
	TWAccountNumber      string            `json:"tw_account_number"`
	Bank                 string            `json:"bank"`
	HasContract          bool              `json:"has_contract"`
	PayrollID            string            `json:"payroll_id"`
	IsCommit             bool              `json:"is_commit"`
	IsPaid               bool              `json:"is_paid"`
	TWGBP                float64           `json:"tw_gbp"` // will be removed
	TWAmount             float64           `json:"tw_amount"`
	TWFee                float64           `json:"tw_fee"`
	TWEmail              string            `json:"tw_email"`
	TWCurrency           string            `json:"tw_currency"`
	Currency             string            `json:"currency"`
}

type payrollBHXHResponse struct {
	DisplayName   string `json:"display_name"`
	BHXH          int64  `json:"bhxh"`
	Batch         int    `json:"batch"`
	AccountNumber string `json:"account_number"`
	Bank          string `json:"bank"`
}

package model

import "time"

type BaseSalary struct {
	BaseModel

	EmployeeID UUID `json:"employee_id"`
	Employee   *Employee

	ContractAmount        int64       `json:"contract_amount"`
	CompanyAccountAmount  int64       `json:"company_account_amount"`
	PersonalAccountAmount int64       `json:"personal_account_amount"`
	InsuranceAmount       VietnamDong `json:"insurance_amount"`
	Type                  string      `json:"type"`
	Category              string      `json:"category"`

	CurrencyID UUID `json:"currency_id"`
	Currency   *Currency
	Batch      int

	EffectiveDate time.Time `json:"effective_date"`
}

func (BaseSalary) TableName() string { return "base_salary" }

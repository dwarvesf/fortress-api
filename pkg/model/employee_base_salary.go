package model

import "time"

type EmployeeBaseSalary struct {
	BaseModel

	EmployeeID     UUID
	CurrencyID     UUID
	StartDate      time.Time
	PayrollBatch   int
	PersonalAmount int64
	ContractAmount int64
	IsActive       bool

	Currency Currency
}
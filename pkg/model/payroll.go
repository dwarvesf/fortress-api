package model

import (
	"encoding/json"
	"fmt"
	"text/template"
	"time"
)

const (
	Payday1st = 1
	Payday2nd = 15

	UnderContract = "Under Contract"
	NonContract   = "Non Contract"
)

// TODO: has not support currencies other than VND yet
type Payroll struct {
	BaseModel

	EmployeeID        UUID
	Total             VietnamDong
	AccountedAmount   VietnamDong
	DueDate           time.Time
	Month             time.Month
	Year              int
	PersonalAmount    int64
	ContractAmount    int64
	Bonus             VietnamDong
	BonusExplain      JSON
	Commission        VietnamDong
	CommissionExplain JSON
	IsPaid            bool
	WiseAmount        float64
	WiseRate          float64
	WiseFee           float64
	Notes             string

	Employee Employee
}

type PayrollExplain struct {
	ID               UUID        `json:"id"`
	Name             string      `json:"name"`
	Month            time.Month  `json:"month"`
	Year             int         `json:"year"`
	Amount           VietnamDong `json:"amount"`
	FormattedAmount  string      `json:"formattedAmount"`
	BasecampTodoID   int64       `json:"todoId"`
	BasecampBucketID int64       `json:"bucketId"`
}

// ToPaidSuccessfulEmailContent to parse the payroll object
// into template when sending email after payroll is paid
func (p Payroll) GetPaidSuccessfulEmailFuncMap(tax float64, Env string) map[string]interface{} {
	// the salary will be the contract(companyAccountAmount in DB)
	// plus the base salary(personalAccountAmount in DB)

	var addresses string
	addresses = "huynh@d.foundation"
	if Env == "prod" {
		addresses = "quang@d.foundation, accounting@d.foundation"
	}

	commissionExplain, bonusExplain := []PayrollExplain{}, []PayrollExplain{}
	json.Unmarshal(p.BonusExplain, &bonusExplain)
	json.Unmarshal(p.CommissionExplain, &commissionExplain)

	return template.FuncMap{
		"ccList": func() string {
			return addresses
		},
		"userFirstName": func() string {
			return getFirstNameFromFullName(p.Employee.FullName)
		},
		"currency": func() string {
			return p.Employee.EmployeeBaseSalary.Currency.Symbol
		},
		"currencyName": func() string {
			return p.Employee.EmployeeBaseSalary.Currency.Name
		},
		"wiseCurrency": func() string {
			return p.Employee.WiseCurrency
		},
		"formattedCurrentMonth": func() string {
			fm := time.Month(int(p.Month))
			return fm.String()
		},
		"formattedBaseSalaryAmount": func() string {
			return formatNumber(p.PersonalAmount)
		},
		"formattedContractAmount": func() string {
			return formatNumber(p.ContractAmount)
		},
		"formattedTotalAllowance": func() string {
			return formatNumber(int64(p.Total))
		},
		"haveBonusOrCommission": func() bool {
			return len(commissionExplain) > 0 || len(bonusExplain) > 0
		},
		"haveCommission": func() bool {
			return len(commissionExplain) > 0
		},
		"haveBonus": func() bool {
			return len(bonusExplain) > 0
		},
		"commissionExplain": func() []PayrollExplain {
			return commissionExplain
		},
		"projectBonusExplains": func() []PayrollExplain {
			return bonusExplain
		},
		"tax": func() string {
			return fmt.Sprintf("%.2f", tax)
		},
	}
}

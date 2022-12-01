package view

import (
	"fmt"
	"sort"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type PayrollResponse struct {
	DisplayName          string            `json:"displayName"`
	BaseSalary           int64             `json:"baseSalary"`
	Bonus                float64           `json:"bonus"`
	TotalWithoutContract float64           `json:"totalWithoutContract"`
	TotalWithContract    model.VietnamDong `json:"totalWithContract"`
	Notes                string            `json:"notes"`
	Date                 int               `json:"date"`
	Month                int               `json:"month"`
	Year                 int               `json:"year"`
	WiseRecipientID      string            `json:"wiseRecipientId"`
	WiseUSD              float64           `json:"wiseUSD"`
	WiseFee              float64           `json:"wiseFee"`
	WiseEmail            string            `json:"wiseEmail"`
	WiseCurrency         string            `json:"wiseCurrency"`
	Currency             string            `json:"currency"`
}

func ToPayrollResponse(p model.Payroll, month time.Month, year int) PayrollResponse {
	bonus := int64(p.Bonus + p.Commission)
	return PayrollResponse{
		DisplayName:          p.Employee.FullName,
		BaseSalary:           p.PersonalAmount,
		Bonus:                float64(bonus),
		TotalWithoutContract: float64(p.PersonalAmount) + float64(bonus),
		TotalWithContract:    p.Total,
		Notes:                p.Notes,
		Date:                 p.Employee.EmployeeBaseSalary.PayrollBatch,
		Month:                int(month),
		Year:                 year,
		WiseRecipientID:      fmt.Sprint(p.Employee.WiseRecipientID),
		WiseFee:              p.WiseFee,
		WiseUSD:              p.WiseAmount,
		WiseEmail:            p.Employee.WiseRecipientEmail,
		WiseCurrency:         p.Employee.WiseCurrency,
		Currency:             p.Employee.EmployeeBaseSalary.Currency.Name,
	}
}

type PayrollDashboardResponse struct {
	EmployeeID               model.UUID `json:"employeeId"`
	Name                     string     `json:"name"`
	Department               string     `json:"department"`
	BaseSalary               int64      `json:"baseSalary"`
	BaseSalaryCurrencySymbol string     `json:"baseSalaryCurrencySymbol"`
	Bonus                    int64      `json:"bonus"`
	BonusCurrencySymbol      string     `json:"bonusCurrencySymbol"`
	Contract                 int64      `json:"contractAmount"`
	ContractCurrencySymbol   string     `json:"contractCurrencySymbol"`
	Allowance                int64      `json:"allowance"`
	AllowanceCurrencySymbol  string     `json:"allowanceCurrencySymbol"`
	WiseEmail                string     `json:"wiseEmail"`
	Bank                     string     `json:"bank"`
}

type lessFunc func(p1, p2 *PayrollDashboardResponse) bool

type multiSorter struct {
	payrolls []PayrollDashboardResponse
	less     []lessFunc
}

func (ms *multiSorter) Len() int {
	return len(ms.payrolls)
}

func (ms *multiSorter) Swap(i, j int) {
	ms.payrolls[i], ms.payrolls[j] = ms.payrolls[j], ms.payrolls[i]
}

func (ms *multiSorter) Less(i, j int) bool {
	p, q := &ms.payrolls[i], &ms.payrolls[j]
	var k int
	for k := range ms.less {
		less := ms.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}
	return ms.less[k](p, q)
}

func (ms *multiSorter) sort(payrolls []PayrollDashboardResponse) {
	ms.payrolls = payrolls
	sort.Sort(ms)
}

func orderedBy(less ...lessFunc) *multiSorter {
	return &multiSorter{less: less}
}

var (
	// ascending order
	nameASC = func(p1, p2 *PayrollDashboardResponse) bool {
		return p1.Name <= p2.Name
	}
	baseSalaryASC = func(p1, p2 *PayrollDashboardResponse) bool {
		return p1.BaseSalary <= p2.BaseSalary
	}
	bonusASC = func(p1, p2 *PayrollDashboardResponse) bool {
		return p1.Bonus <= p2.Bonus
	}
	contractASC = func(p1, p2 *PayrollDashboardResponse) bool {
		return p1.Contract <= p2.Contract
	}
	allowanceASC = func(p1, p2 *PayrollDashboardResponse) bool {
		return p1.Allowance <= p2.Allowance
	}
	bankASC = func(p1, p2 *PayrollDashboardResponse) bool {
		return p1.Bank <= p2.Bank
	}

	// descending order
	nameDESC = func(p1, p2 *PayrollDashboardResponse) bool {
		return p1.Name > p2.Name
	}
	baseSalaryDESC = func(p1, p2 *PayrollDashboardResponse) bool {
		return p1.BaseSalary > p2.BaseSalary
	}
	bonusDESC = func(p1, p2 *PayrollDashboardResponse) bool {
		return p1.Bonus > p2.Bonus
	}
	contractDESC = func(p1, p2 *PayrollDashboardResponse) bool {
		return p1.Contract > p2.Contract
	}
	allowanceDESC = func(p1, p2 *PayrollDashboardResponse) bool {
		return p1.Allowance > p2.Allowance
	}
	bankDESC = func(p1, p2 *PayrollDashboardResponse) bool {
		return p1.Bank > p2.Bank
	}
)

func SortDashboardPayrolls(payrolls []PayrollDashboardResponse, sort map[string]string) []PayrollDashboardResponse {
	switch {
	// name
	case sort["name"] == "asc":
		orderedBy(nameASC).sort(payrolls)
	case sort["name"] == "desc":
		orderedBy(nameDESC).sort(payrolls)

	// base salary
	case sort["base_salary"] == "asc":
		orderedBy(baseSalaryASC).sort(payrolls)
	case sort["base_salary"] == "desc":
		orderedBy(baseSalaryDESC).sort(payrolls)

	// bonus
	case sort["bonus"] == "asc":
		orderedBy(bonusASC).sort(payrolls)
	case sort["bonus"] == "desc":
		orderedBy(bonusDESC).sort(payrolls)

	// contract
	case sort["contract"] == "asc":
		orderedBy(contractASC).sort(payrolls)
	case sort["contract"] == "desc":
		orderedBy(contractDESC).sort(payrolls)

	// allowance
	case sort["allowance"] == "asc":
		orderedBy(allowanceASC).sort(payrolls)
	case sort["allowance"] == "desc":
		orderedBy(allowanceDESC).sort(payrolls)

	// bank
	case sort["bank"] == "asc":
		orderedBy(bankASC).sort(payrolls)
	case sort["bank"] == "desc":
		orderedBy(bankDESC).sort(payrolls)

	default:
	}
	return payrolls
}

type PayrollList struct {
	Payrolls   []PayrollResponse `json:"payrolls"`
	SubTotal   int64             `json:"subTotal"`
	BonusTotal int64             `json:"bonusTotal"`
}

func ToPayrollList(payrolls []model.Payroll, month time.Month, year int) *PayrollList {
	rs := make([]PayrollResponse, 0, len(payrolls))
	var subTotal int64
	for _, p := range payrolls {
		r := ToPayrollResponse(p, month, year)
		subTotal += int64(r.TotalWithoutContract)
		rs = append(rs, r)
	}
	return &PayrollList{
		Payrolls: rs,
		SubTotal: subTotal,
	}
}

type BhxhResponse struct {
	DisplayName   string `json:"displayName"`
	BHXH          int64  `json:"bhxh"`
	Batch         int    `json:"batch"`
	AccountNumber string `json:"accountNumber"`
	Bank          string `json:"bank"`
}

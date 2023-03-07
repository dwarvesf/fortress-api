package model

import "time"

// EmployeeCommission represents EmployeeCommission table
// save the commission when invoice is paid for an Employee
type EmployeeCommission struct {
	BaseModel
	EmployeeID     UUID        `json:"employee_id"`
	InvoiceID      UUID        `json:"invoice_id"`
	Project        string      `json:"project"`
	IsPaid         bool        `json:"is_paid"`
	Amount         VietnamDong `json:"amount"`
	ConversionRate float64
	Formula        string    `json:"string"`
	Note           string    `json:"note"`
	PaidAt         time.Time `json:"paid_at"`

	Employee *Employee `json:"-" gorm:"foreignkey:EmployeeID"`
	Invoice  *Invoice  `json:"-" gorm:"foreignkey:InvoiceID"`
}

func (EmployeeCommission) TableName() string { return "employee_commissions" }

// NewEmployeeCommission create new Employee commission
func NewEmployeeCommission(employeeID, invoiceID UUID, project string, amount VietnamDong, rate float64) EmployeeCommission {
	return EmployeeCommission{
		EmployeeID:     employeeID,
		InvoiceID:      invoiceID,
		Amount:         amount,
		Project:        project,
		ConversionRate: rate,
	}
}

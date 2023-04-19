package model

import "time"

// EmployeeCommission represents EmployeeCommission table
// save the commission when invoice is paid for an Employee
type EmployeeCommission struct {
	BaseModel

	EmployeeID     UUID
	InvoiceID      UUID
	Project        string
	IsPaid         bool
	Amount         VietnamDong
	ConversionRate float64
	Formula        string
	Note           string
	PaidAt         *time.Time

	Employee *Employee
	Invoice  *Invoice
}

// New create new Employee commission
func New(employeeID, invoiceID UUID, project string, amount VietnamDong, rate float64) EmployeeCommission {
	return EmployeeCommission{
		EmployeeID:     employeeID,
		InvoiceID:      invoiceID,
		Amount:         amount,
		Project:        project,
		ConversionRate: rate,
	}
}

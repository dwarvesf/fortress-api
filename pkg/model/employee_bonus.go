package model

// EmployeeBonus represents employeeBonus table
// save the commission when invoice is paid for an employee
type EmployeeBonus struct {
	ID         UUID        `json:"id"`
	EmployeeID UUID        `json:"employee_id"`
	Amount     VietnamDong `json:"amount"`
	IsActive   bool        `json:"is_active"`
	Name       string      `json:"name"`
}

func (EmployeeBonus) TableName() string { return "employee_bonuses" }

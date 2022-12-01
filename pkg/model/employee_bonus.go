package model

type EmployeeBonus struct {
	BaseModel

	EmployeeID UUID
	Amount     VietnamDong
	Name       string
	IsActive   bool
}

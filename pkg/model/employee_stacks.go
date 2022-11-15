package model

type EmployeeStack struct {
	BaseModel

	EmployeeID UUID
	StackID    UUID

	Stack Stack
}

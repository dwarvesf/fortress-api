package model

type EmployeeRole struct {
	BaseModel

	EmployeeID string
	RoleID     string

	Role Role
}

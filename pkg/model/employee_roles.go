package model

type EmployeeRole struct {
	BaseModel

	EmployeeID UUID
	RoleID     UUID

	Role Role
}

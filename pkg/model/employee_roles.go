package model

type EmployeeRoles struct {
	BaseModel
	EmployeeID UUID
	RoleID     UUID

	Role *Role `gorm:"foreignkey:role_id"`
}

package model

type EmployeePosition struct {
	BaseModel

	EmployeeID string
	PositionID string

	Position Position `gorm:"foreignkey:position_id"`
}

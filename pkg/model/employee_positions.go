package model

type EmployeePosition struct {
	BaseModel

	EmployeeID UUID
	PositionID UUID

	Position Position
}

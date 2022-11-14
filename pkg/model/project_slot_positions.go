package model

type ProjectSlotPosition struct {
	BaseModel

	ProjectSlotID UUID
	PositionID    UUID

	Position Position
}

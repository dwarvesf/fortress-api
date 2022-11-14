package model

type ProjectMemberPosition struct {
	BaseModel

	ProjectMemberID UUID
	PositionID      UUID

	Position Position
}

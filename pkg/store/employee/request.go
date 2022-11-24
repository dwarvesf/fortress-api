package employee

type GetAllInput struct {
	WorkingStatus string
	Preload       bool
	PositionID    string
	StackID       string
	ProjectID     string
	Keyword       string
}

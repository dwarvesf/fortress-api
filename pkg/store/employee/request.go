package employee

type GetAllInput struct {
	WorkingStatuses []string
	Preload         bool
	PositionID      string
	StackID         string
	ProjectID       string
	Keyword         string
}

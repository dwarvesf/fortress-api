package employee

type EmployeeFilter struct {
	WorkingStatuses []string
	Preload         bool
	PositionID      string
	StackID         string
	ProjectID       string
	Keyword         string
	//field sort
	JoinedDateSort string
}

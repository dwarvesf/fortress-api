package employee

import "github.com/dwarvesf/fortress-api/pkg/model"

type EmployeeFilter struct {
	WorkingStatuses []string
	Preload         bool
	PositionID      string
	StackID         string
	ProjectID       string
	Keyword         string
	//field sort
	JoinedDateSort model.SortOrder
}

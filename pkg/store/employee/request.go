package employee

import "github.com/dwarvesf/fortress-api/pkg/model"

type EmployeeFilter struct {
	WorkingStatuses []string
	Preload         bool
	Positions       []string
	Stacks          []string
	Projects        []string
	Chapters        []string
	Seniorities     []string
	Organizations   []string
	LineManagers    []string
	Keyword         string
	//field sort
	JoinedDateSort model.SortOrder
}

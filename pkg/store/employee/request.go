package employee

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

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
	IsLeft          *bool
	BatchDate       *time.Time
	//field sort
	JoinedDateSort model.SortOrder
}

type DiscordRequestFilter struct {
	DiscordID []string
	Email     string
	Github    string
}

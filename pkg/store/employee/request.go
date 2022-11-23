package employee

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// TODO: sao file này đặt tên là request.go còn project lại là project/filter.go
type SearchFilter struct {
	EmployeeID    string
	WorkingStatus string
	Preload       bool
	PositionID    string
	StackID       string
	ProjectID     string
	Keyword       string
}

type UpdateGeneralInfoInput struct {
	FullName      string
	Email         string
	Phone         string
	LineManagerID model.UUID
	DiscordID     string
	GithubID      string
	NotionID      string
}

type UpdateSkillsInput struct {
	Positions []model.UUID
	Chapter   model.UUID
	Seniority model.UUID
	Stacks    []model.UUID
}

type UpdatePersonalInfoInput struct {
	DoB           *time.Time
	Gender        string
	Address       string
	PersonalEmail string
}

type UpdateProfileInforInput struct {
	PersonalEmail string
	TeamEmail     string
	PhoneNumber   string
	DiscordID     string
	GithubID      string
	NotionID      string
}

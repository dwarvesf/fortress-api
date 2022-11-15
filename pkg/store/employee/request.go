package employee

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type SearchFilter struct {
	WorkingStatus string
}

type UpdateGeneralInfoInput struct {
	FullName      string
	Email         string
	Phone         string
	LineManagerID string
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

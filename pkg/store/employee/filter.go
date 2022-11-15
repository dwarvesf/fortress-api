package employee

import "github.com/dwarvesf/fortress-api/pkg/model"

type SearchFilter struct {
	WorkingStatus string
}

type EditGeneralInfoInput struct {
	FullName      string
	Email         string
	Phone         string
	LineManagerID string
	DiscordID     string
	GithubID      string
	NotionID      string
}

type EditSkillsInput struct {
	Positions []model.UUID
	Chapter   model.UUID
	Seniority model.UUID
	Stacks    []model.UUID
}

package employee

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type GetListEmployeeQuery struct {
	model.Pagination

	WorkingStatus string `json:"workingStatus" form:"workingStatus"`
}

type UpdateGeneralInfoInput struct {
	Fullname      string     `form:"fullName" json:"fullName" binding:"required,max=99"`
	Email         string     `form:"email" json:"email" binding:"required,email"`
	Phone         string     `form:"phone" json:"phone" binding:"required,max=12,min=10"`
	LineManagerID model.UUID `form:"lineManagerID" json:"lineManagerID"`
	DiscordID     string     `form:"discordID" json:"discordID"`
	GithubID      string     `form:"githubID" json:"githubID"`
	NotionID      string     `form:"notionID" json:"notionID"`
}

// CreateEmployee view for create new employee
type CreateEmployee struct {
	FullName      string     `json:"fullName" binding:"required,max=100"`
	DisplayName   string     `json:"displayName" binding:"required"`
	TeamEmail     string     `json:"teamEmail" binding:"required,email"`
	PersonalEmail string     `json:"personalEmail" binding:"required,email"`
	PositionID    model.UUID `json:"positionID" binding:"required"`
	Salary        int        `json:"salary" binding:"required"`
	SeniorityID   model.UUID `json:"seniorityID" binding:"required"`
	RoleID        model.UUID `json:"roleID" binding:"required"`
}

type UpdateSkillsInput struct {
	Positions []model.UUID `form:"positions" json:"positions" binding:"required"`
	Chapter   model.UUID   `form:"chapter" json:"chapter"`
	Seniority model.UUID   `form:"seniority" json:"seniority" binding:"required"`
	Stacks    []model.UUID `form:"stacks" json:"stacks" binding:"required"`
}

type UpdatePersonalInfoInput struct {
	DoB           *time.Time `form:"dob" json:"dob" binding:"required"`
	Gender        string     `form:"gender" json:"gender" binding:"required"`
	Address       string     `form:"address" json:"address" binding:"required,max=200"`
	PersonalEmail string     `form:"personalEmail" json:"personalEmail" binding:"required,email"`
}

package employee

import "github.com/dwarvesf/fortress-api/pkg/model"

type GetListEmployeeQuery struct {
	model.Pagination

	WorkingStatus string `json:"workingStatus" form:"workingStatus"`
}

type EditGeneralInfo struct {
	Fullname      string `form:"fullName" json:"fullName" binding:"required,max=99"`
	Email         string `form:"email" json:"email" binding:"required,email"`
	Phone         string `form:"phone" json:"phone" binding:"required,max=12,min=10"`
	LineManagerID string `form:"lineManagerID" json:"lineManagerID"`
	DiscordID     string `form:"discordID" json:"discordID"`
	GithubID      string `form:"githubID" json:"githubID"`
	NotionID      string `form:"notionID" json:"notionID"`
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

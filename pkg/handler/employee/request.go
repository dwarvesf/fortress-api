package employee

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type GetListEmployeeQuery struct {
	model.Pagination

	WorkingStatus string `json:"workingStatus" form:"workingStatus"`
	Preload       bool   `json:"preload" form:"preload,default=true"`
}

type UpdateGeneralInfoInput struct {
	FullName      string     `form:"fullName" json:"fullName" binding:"required,max=99"`
	Email         string     `form:"email" json:"email" binding:"required,email"`
	Phone         string     `form:"phone" json:"phone" binding:"required,max=12,min=10"`
	LineManagerID model.UUID `form:"lineManagerID" json:"lineManagerID"`
	DiscordID     string     `form:"discordID" json:"discordID"`
	GithubID      string     `form:"githubID" json:"githubID"`
	NotionID      string     `form:"notionID" json:"notionID"`
}

// CreateEmployeeInput view for create new employee
type CreateEmployeeInput struct {
	FullName      string       `json:"fullName" binding:"required,max=100"`
	DisplayName   string       `json:"displayName"`
	TeamEmail     string       `json:"teamEmail" binding:"required,email"`
	PersonalEmail string       `json:"personalEmail" binding:"required,email"`
	Positions     []model.UUID `form:"positions" json:"positions" binding:"required"`
	Salary        int          `json:"salary" binding:"required"`
	SeniorityID   model.UUID   `json:"seniorityID" binding:"required"`
	RoleID        model.UUID   `json:"roleID" binding:"required"`
	Status        string       `json:"status" binding:"required"`
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

type UpdateWorkingStatusInput struct {
	EmployeeStatus model.WorkingStatus `json:"employeeStatus"`
}

func (i *UpdateWorkingStatusInput) Validate() error {
	if !model.WorkingStatus(i.EmployeeStatus).IsValid() {
		return ErrInvalidEmployeeStatus
	}

	return nil
}

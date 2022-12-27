package request

import (
	"regexp"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/handler/employee/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type GetListEmployeeInput struct {
	model.Pagination `json:"pagination" form:"pagination"`

	WorkingStatuses []string `json:"workingStatuses" form:"workingStatuses"`
	Preload         bool     `json:"preload" form:"preload,default=true"`
	Positions       []string `json:"positions" form:"positions"`
	Stacks          []string `json:"stacks" form:"stacks"`
	Projects        []string `json:"projects" form:"projects"`
	Chapters        []string `json:"chapters" form:"chapters"`
	Seniorities     []string `json:"seniorities" form:"seniorities"`
	Keyword         string   `json:"keyword" form:"keyword"`
}

type UpdateEmployeeGeneralInfoInput struct {
	FullName      string     `form:"fullName" json:"fullName" binding:"required,max=99"`
	Email         string     `form:"email" json:"email" binding:"required,email"`
	Phone         string     `form:"phone" json:"phone" binding:"required,max=12,min=10"`
	LineManagerID model.UUID `form:"lineManagerID" json:"lineManagerID"`
	GithubID      string     `form:"githubID" json:"githubID"`
	NotionID      string     `form:"notionID" json:"notionID"`
	NotionName    string     `form:"notionName" json:"notionName"`
	DiscordID     string     `form:"discordID" json:"discordID"`
	DiscordName   string     `form:"discordName" json:"discordName"`
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
	Positions       []model.UUID `form:"positions" json:"positions" binding:"required"`
	LeadingChapters []model.UUID `form:"leadingChapters" json:"leadingChapters"`
	Chapters        []model.UUID `form:"chapters" json:"chapters" binding:"required"`
	Seniority       model.UUID   `form:"seniority" json:"seniority" binding:"required"`
	Stacks          []model.UUID `form:"stacks" json:"stacks" binding:"required"`
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
	if !i.EmployeeStatus.IsValid() {
		return errs.ErrInvalidEmployeeStatus
	}

	return nil
}

func (input *GetListEmployeeInput) Validate() error {
	if len(input.Positions) > 0 {
		for _, p := range input.Positions {
			if p == "" {
				return errs.ErrInvalidPositionCode
			}
		}
	}
	if len(input.Stacks) > 0 {
		for _, s := range input.Stacks {
			if s == "" {
				return errs.ErrInvalidStackCode
			}
		}
	}
	if len(input.Projects) > 0 {
		for _, p := range input.Projects {
			if p == "" {
				return errs.ErrInvalidProjectCode
			}
		}
	}
	if len(input.Chapters) > 0 {
		for _, c := range input.Chapters {
			if c == "" {
				return errs.ErrInvalidChapterCode
			}
		}
	}
	if len(input.Seniorities) > 0 {
		for _, s := range input.Seniorities {
			if s == "" {
				return errs.ErrInvalidSeniorityCode
			}
		}
	}

	return nil
}

func (input CreateEmployeeInput) Validate() error {
	regex, _ := regexp.Compile(".+@((dwarvesv\\.com)|(d\\.foundation))")
	if !regex.MatchString(input.TeamEmail) {
		return errs.ErrInvalidEmailDomain
	}

	if !model.WorkingStatus(input.Status).IsValid() {
		return errs.ErrInvalidEmployeeStatus
	}

	return nil
}

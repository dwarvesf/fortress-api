package request

import (
	"regexp"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/handler/employee/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type GetListEmployeeInput struct {
	model.Pagination

	WorkingStatuses []string `json:"workingStatuses" form:"workingStatuses"`
	Preload         bool     `json:"preload" form:"preload,default=true"`
	Positions       []string `json:"positions" form:"positions"`
	Stacks          []string `json:"stacks" form:"stacks"`
	Projects        []string `json:"projects" form:"projects"`
	Chapters        []string `json:"chapters" form:"chapters"`
	Seniorities     []string `json:"seniorities" form:"seniorities"`
	Organizations   []string `json:"organizations" form:"organizations"`
	LineManagers    []string `json:"lineManagers" form:"lineManagers"`
	Keyword         string   `json:"keyword" form:"keyword"`
}

type UpdateEmployeeGeneralInfoInput struct {
	FullName        string       `form:"fullName" json:"fullName" binding:"required,max=99"`
	Email           string       `form:"email" json:"email" binding:"required,email"`
	Phone           string       `form:"phone" json:"phone" binding:"required,max=18,min=9"`
	LineManagerID   model.UUID   `form:"lineManagerID" json:"lineManagerID"`
	DisplayName     string       `form:"displayName" json:"displayName"`
	GithubID        string       `form:"githubID" json:"githubID"`
	NotionID        string       `form:"notionID" json:"notionID"`
	NotionName      string       `form:"notionName" json:"notionName"`
	NotionEmail     string       `form:"notionEmail" json:"notionEmail"`
	DiscordID       string       `form:"discordID" json:"discordID"`
	DiscordName     string       `form:"discordName" json:"discordName"`
	LinkedInName    string       `form:"linkedInName" json:"linkedInName"`
	LeftDate        string       `form:"leftDate" json:"leftDate"`
	JoinedDate      string       `form:"joinedDate" json:"joinedDate"`
	OrganizationIDs []model.UUID `form:"organizationIDs" json:"organizationIDs"`
	ReferredBy      model.UUID   `form:"referredBy" json:"referredBy"`
}

type AddMenteeInput struct {
	MenteeID model.UUID `form:"menteeID" json:"menteeID" binding:"required"`
}

type DeleteMenteeInput struct {
	MentorID string
	MenteeID string
}

func (e *DeleteMenteeInput) Validate() error {
	if e.MentorID == "" || !model.IsUUIDFromString(e.MentorID) {
		return errs.ErrInvalidEmployeeID
	}

	if e.MenteeID == "" || !model.IsUUIDFromString(e.MenteeID) {
		return errs.ErrInvalidEmployeeID
	}

	return nil
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
	Roles         []model.UUID `json:"roles" binding:"required"`
	Status        string       `json:"status" binding:"required"`
	ReferredBy    model.UUID   `json:"referredBy"`
}

type UpdateSkillsInput struct {
	Positions       []model.UUID `form:"positions" json:"positions" binding:"required"`
	LeadingChapters []model.UUID `form:"leadingChapters" json:"leadingChapters"`
	Chapters        []model.UUID `form:"chapters" json:"chapters" binding:"required"`
	Seniority       model.UUID   `form:"seniority" json:"seniority" binding:"required"`
	Stacks          []model.UUID `form:"stacks" json:"stacks" binding:"required"`
}

type UpdatePersonalInfoInput struct {
	DoB              *time.Time `form:"dob" json:"dob" binding:"required"`
	Gender           string     `form:"gender" json:"gender" binding:"required"`
	PlaceOfResidence string     `form:"placeOfResidence" json:"placeOfResidence"`
	Address          string     `form:"address" json:"address" binding:"required,max=200"`
	PersonalEmail    string     `form:"personalEmail" json:"personalEmail" binding:"required,email"`
	Country          string     `form:"country" json:"country"`
	City             string     `form:"city" json:"city"`
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
			if strings.TrimSpace(p) == "" {
				return errs.ErrInvalidPositionCode
			}
		}
	}
	if len(input.Stacks) > 0 {
		for _, s := range input.Stacks {
			if strings.TrimSpace(s) == "" {
				return errs.ErrInvalidStackCode
			}
		}
	}
	if len(input.Projects) > 0 {
		for _, p := range input.Projects {
			if strings.TrimSpace(p) == "" {
				return errs.ErrInvalidProjectCode
			}
		}
	}
	if len(input.Chapters) > 0 {
		for _, c := range input.Chapters {
			if strings.TrimSpace(c) == "" {
				return errs.ErrInvalidChapterCode
			}
		}
	}
	if len(input.Seniorities) > 0 {
		for _, s := range input.Seniorities {
			if strings.TrimSpace(s) == "" {
				return errs.ErrInvalidSeniorityCode
			}
		}
	}
	if len(input.Organizations) > 0 {
		for _, v := range input.Organizations {
			if strings.TrimSpace(v) == "" {
				return errs.ErrInvalidOrganizationCode
			}
		}
	}

	return nil
}

func (input CreateEmployeeInput) Validate() error {
	regex, _ := regexp.Compile(".+@((dwarvesv\\.com)|(d\\.foundation)|(gmail\\.com))")
	if !regex.MatchString(input.TeamEmail) {
		return errs.ErrInvalidEmailDomain
	}

	if !model.WorkingStatus(input.Status).IsValid() {
		return errs.ErrInvalidEmployeeStatus
	}

	if len(input.Roles) == 0 {
		return errs.ErrRoleCannotBeEmpty
	}

	return nil
}

type UpdateRoleBody struct {
	Roles []model.UUID `form:"roles" json:"roles" binding:"required"`
}

type UpdateRoleInput struct {
	EmployeeID string
	Body       UpdateRoleBody
}

func (i UpdateRoleInput) Validate() error {
	if i.EmployeeID == "" || !model.IsUUIDFromString(i.EmployeeID) {
		return errs.ErrInvalidEmployeeID
	}

	if len(i.Body.Roles) == 0 {
		return errs.ErrRoleCannotBeEmpty
	}

	return nil
}

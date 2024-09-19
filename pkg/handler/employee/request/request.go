package request

import (
	"regexp"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/handler/employee/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type GetListEmployeeQuery struct {
	view.Pagination

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
} // @name GetListEmployeeQuery

type UpdateEmployeeGeneralInfoRequest struct {
	FullName           string      `form:"fullName" json:"fullName" binding:"required,max=99"`
	Email              string      `form:"email" json:"email" binding:"required,email"`
	Phone              string      `form:"phone" json:"phone" binding:"required,max=18,min=9"`
	LineManagerID      view.UUID   `form:"lineManagerID" json:"lineManagerID"`
	DisplayName        string      `form:"displayName" json:"displayName"`
	GithubID           string      `form:"githubID" json:"githubID"`
	NotionID           string      `form:"notionID" json:"notionID"`
	NotionName         string      `form:"notionName" json:"notionName"`
	NotionEmail        string      `form:"notionEmail" json:"notionEmail"`
	DiscordID          string      `form:"discordID" json:"discordID"`
	DiscordName        string      `form:"discordName" json:"discordName"`
	LinkedInName       string      `form:"linkedInName" json:"linkedInName"`
	LeftDate           string      `form:"leftDate" json:"leftDate"`
	JoinedDate         string      `form:"joinedDate" json:"joinedDate"`
	OrganizationIDs    []view.UUID `form:"organizationIDs" json:"organizationIDs"`
	ReferredBy         view.UUID   `form:"referredBy" json:"referredBy"`
	WiseRecipientID    string      `form:"wiseRecipientID" json:"wiseRecipientID"`
	WiseRecipientEmail string      `form:"wiseRecipientEmail" json:"wiseRecipientEmail"`
	WiseRecipientName  string      `form:"wiseRecipientName" json:"wiseRecipientName"`
	WiseAccountNumber  string      `form:"wiseAccountNumber" json:"wiseAccountNumber"`
	WiseCurrency       string      `form:"wiseCurrency" json:"wiseCurrency"`
} // @name UpdateEmployeeGeneralInfoRequest
type UpdateBaseSalaryRequest struct {
	ContractAmount        int64      `form:"contractAmount" json:"contractAmount" binding:"gte=0"`
	CompanyAccountAmount  int64      `form:"companyAccountAmount" json:"companyAccountAmount" binding:"gte=0"`
	PersonalAccountAmount int64      `form:"personalAccountAmount" json:"personalAccountAmount" binding:"gte=0"`
	CurrencyCode          string     `form:"currencyCode" json:"currencyCode" binding:"required"`
	EffectiveDate         *time.Time `form:"effectiveDate" json:"effectiveDate"`
	Batch                 int        `form:"batch" json:"batch" binding:"required,eq=1|eq=15"`
} // @name UpdateBaseSalaryRequest

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

// CreateEmployeeRequest view for create new employee
type CreateEmployeeRequest struct {
	FullName      string       `json:"fullName" binding:"required,max=100"`
	DisplayName   string       `json:"displayName" binding:"required"`
	TeamEmail     string       `json:"teamEmail" binding:"required"`
	PersonalEmail string       `json:"personalEmail" binding:"required,email"`
	Positions     []model.UUID `form:"positions" json:"positions" binding:"required"`
	Salary        int64        `json:"salary" binding:"required"`
	SeniorityID   model.UUID   `json:"seniorityID" binding:"required"`
	Roles         []model.UUID `json:"roles" binding:"required"`
	Status        string       `json:"status" binding:"required"`
	ReferredBy    model.UUID   `json:"referredBy"`
	JoinedDate    string       `json:"joinedDate" binding:"required"`
} // @name CreateEmployeeRequest

type UpdateSkillsRequest struct {
	Positions       []model.UUID `form:"positions" json:"positions" binding:"required"`
	LeadingChapters []model.UUID `form:"leadingChapters" json:"leadingChapters"`
	Chapters        []model.UUID `form:"chapters" json:"chapters" binding:"required"`
	Seniority       model.UUID   `form:"seniority" json:"seniority" binding:"required"`
	Stacks          []model.UUID `form:"stacks" json:"stacks" binding:"required"`
} // @name UpdateSkillsRequest

type UpdatePersonalInfoRequest struct {
	DoB              *time.Time `form:"dob" json:"dob" binding:"required"`
	Gender           string     `form:"gender" json:"gender" binding:"required"`
	PlaceOfResidence string     `form:"placeOfResidence" json:"placeOfResidence"`
	Address          string     `form:"address" json:"address" binding:"required,max=200"`
	PersonalEmail    string     `form:"personalEmail" json:"personalEmail" binding:"required,email"`
	Country          string     `form:"country" json:"country" binding:"required"`
	City             string     `form:"city" json:"city" binding:"required"`
} // @name UpdatePersonalInfoRequest

type UpdateWorkingStatusRequest struct {
	EmployeeStatus WorkingStatus `json:"employeeStatus"`
} // @name UpdateWorkingStatusRequest

type WorkingStatus string // @name WorkingStatus

const (
	WorkingStatusOnBoarding WorkingStatus = "on-boarding"
	WorkingStatusLeft       WorkingStatus = "left"
	WorkingStatusProbation  WorkingStatus = "probation"
	WorkingStatusFullTime   WorkingStatus = "full-time"
	WorkingStatusContractor WorkingStatus = "contractor"
)

func (e WorkingStatus) IsValid() bool {
	switch e {
	case
		WorkingStatusOnBoarding,
		WorkingStatusContractor,
		WorkingStatusLeft,
		WorkingStatusProbation,
		WorkingStatusFullTime:
		return true
	}
	return false
}

// String returns the string type from the WorkingStatus type
func (e WorkingStatus) String() string {
	return string(e)
}

func (i *UpdateWorkingStatusRequest) Validate() error {
	if !i.EmployeeStatus.IsValid() {
		return errs.ErrInvalidEmployeeStatus
	}

	return nil
}

func (input *GetListEmployeeQuery) Validate() error {
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

func (i *CreateEmployeeRequest) Validate() error {
	teamEmailRegex := ".+@((dwarvesv\\.com)|(d\\.foundation))"
	regex, _ := regexp.Compile(teamEmailRegex)
	if i.TeamEmail != "" && !regex.MatchString(i.TeamEmail) {
		return errs.ErrInvalidEmailDomain
	}

	if !model.WorkingStatus(i.Status).IsValid() {
		return errs.ErrInvalidEmployeeStatus
	}

	if len(i.Roles) == 0 {
		return errs.ErrRoleCannotBeEmpty
	}

	_, err := time.Parse("2006-01-02", i.JoinedDate)
	if i.JoinedDate != "" && err != nil {
		return errs.ErrInvalidJoinedDate
	}

	return nil
}

func (i *CreateEmployeeRequest) GetJoinedDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.JoinedDate)
	if i.JoinedDate == "" || err != nil {
		return nil
	}

	return &date
}

type UpdateRoleRequest struct {
	Roles []model.UUID `form:"roles" json:"roles" binding:"required"`
} // @name UpdateRoleRequest

type UpdateRoleInput struct {
	EmployeeID string
	Body       UpdateRoleRequest
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

type SalaryAdvanceRequest struct {
	DiscordID string `json:"discordID"`
	Amount    string `json:"amount"`
} // @name SalaryAdvanceRequest

type SalaryAdvanceReportRequest struct {
	view.Pagination
	model.SortOrder `json:"sortOrder" form:"sortOrder"`

	IsPaid *bool `json:"isPaid" form:"isPaid"`
} // @name SalaryAdvanceReportRequest

type GetEmployeeEarnTransactionsRequest struct {
	view.Pagination
} // @name GetEmployeeEarnTransactionsRequest

type CheckInRequest struct {
	CheckIns []CheckIn `json:"check_ins" binding:"required,dive,required"`
} // @name CheckInRequest

type CheckIn struct {
	DiscordID string    `json:"discord_id" binding:"required"`
	Time      time.Time `json:"time" binding:"required"`
}

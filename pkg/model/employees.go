package model

import "time"

// WorkingStatus working_status type for employee table
type WorkingStatus string

// values for working_status
const (
	WorkingStatusOnBoarding WorkingStatus = "on-boarding"
	WorkingStatusLeft       WorkingStatus = "left"
	WorkingStatusProbation  WorkingStatus = "probation"
	WorkingStatusFullTime   WorkingStatus = "full-time"
	WorkingStatusContractor WorkingStatus = "contractor"
)

// IsValid validation for WorkingStatus
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

// AccountRole account_role type for employee table
type AccountRole string

// values for account_role
const (
	AccountRoleAdmin  AccountRole = "admin"
	AccountRoleMember AccountRole = "member"
)

// IsValid validation for AccountRole
func (e AccountRole) IsValid() bool {
	switch e {
	case
		AccountRoleAdmin,
		AccountRoleMember:
		return true
	}
	return false
}

// Employee define the model for table employees
type Employee struct {
	BaseModel

	// basic info
	FullName               string
	DisplayName            string
	Username               string
	TeamEmail              string
	PersonalEmail          string
	Avatar                 string
	PhoneNumber            string
	Address                string
	ShelterAddress         string
	PermanentAddress       string
	MBTI                   string
	Gender                 string
	Horoscope              string
	PassportPhotoFront     string
	PassportPhotoBack      string
	IdentityCardPhotoFront string
	IdentityCardPhotoBack  string
	DateOfBirth            *time.Time
	Country                string
	City                   string

	// working info
	WorkingStatus WorkingStatus
	JoinedDate    *time.Time
	LeftDate      *time.Time
	SeniorityID   UUID
	LineManagerID UUID

	// social services
	BasecampID             string
	BasecampAttachableSGID string `gorm:"column:basecamp_attachable_sgid"`
	GitlabID               string
	GithubID               string
	NotionID               string
	NotionName             string
	NotionEmail            string
	DiscordID              string
	DiscordName            string
	LinkedInName           string `gorm:"column:linkedin_name"`

	// payroll info
	WiseRecipientEmail string
	WiseRecipientID    string
	WiseRecipientName  string
	WiseAccountNumber  string
	WiseCurrency       string

	LocalBankBranch        string
	LocalBankNumber        string
	LocalBankCurrency      string
	LocalBranchName        string
	LocalBankRecipientName string

	Seniority         *Seniority
	LineManager       *Employee
	ProjectMembers    []ProjectMember
	Roles             []Role     `gorm:"many2many:employee_roles;"`
	Positions         []Position `gorm:"many2many:employee_positions;"`
	EmployeeRoles     []EmployeeRole
	EmployeePositions []EmployeePosition
	EmployeeStacks    []EmployeeStack
	EmployeeChapters  []EmployeeChapter
}

// ToEmployeeMap create map from employees
func ToEmployeeMap(employees []*Employee) map[UUID]Employee {
	rs := map[UUID]Employee{}
	for _, e := range employees {
		rs[e.ID] = *e
	}

	return rs
}

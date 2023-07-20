package model

import (
	"strings"
	"time"
)

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

// String returns the string representation
func (e AccountRole) String() string {
	return string(e)
}

// Employee define the model for table employees
type Employee struct {
	BaseModel

	// basic info
	FullName               string     `gorm:"default:null"`
	DisplayName            string     `gorm:"default:null"`
	Username               string     `gorm:"default:null"`
	TeamEmail              string     `gorm:"default:null"`
	PersonalEmail          string     `gorm:"default:null"`
	Avatar                 string     `gorm:"default:null"`
	PhoneNumber            string     `gorm:"default:null"`
	Address                string     `gorm:"default:null"`
	PlaceOfResidence       string     `gorm:"default:null"`
	MBTI                   string     `gorm:"default:null"`
	Gender                 string     `gorm:"default:null"`
	Horoscope              string     `gorm:"default:null"`
	PassportPhotoFront     string     `gorm:"default:null"`
	PassportPhotoBack      string     `gorm:"default:null"`
	IdentityCardPhotoFront string     `gorm:"default:null"`
	IdentityCardPhotoBack  string     `gorm:"default:null"`
	DateOfBirth            *time.Time `gorm:"default:null"`
	Country                string     `gorm:"default:null"`
	City                   string     `gorm:"default:null"`
	Lat                    string     `gorm:"default:null"`
	Long                   string     `gorm:"default:null"`

	// working info
	WorkingStatus WorkingStatus `gorm:"default:null"`
	JoinedDate    *time.Time    `gorm:"default:null"`
	LeftDate      *time.Time    `gorm:"default:null"`
	SeniorityID   UUID          `gorm:"default:null"`
	LineManagerID UUID          `gorm:"default:null"`
	ReferredBy    UUID          `gorm:"default:null"`

	// social services
	BasecampID             int    `gorm:"default:null"`
	BasecampAttachableSGID string `gorm:"column:basecamp_attachable_sgid;default:null"`
	DiscordAccountID       UUID   `gorm:"default:null"`

	// payroll info
	WiseRecipientEmail string `gorm:"default:null"`
	WiseRecipientID    string `gorm:"default:null"`
	WiseRecipientName  string `gorm:"default:null"`
	WiseAccountNumber  string `gorm:"default:null"`
	WiseCurrency       string `gorm:"default:null"`

	LocalBankBranch        string `gorm:"default:null"`
	LocalBankNumber        string `gorm:"default:null"`
	LocalBankCurrency      string `gorm:"default:null"`
	LocalBranchName        string `gorm:"default:null"`
	LocalBankRecipientName string `gorm:"default:null"`

	Seniority             *Seniority
	LineManager           *Employee
	Referrer              *Employee `gorm:"foreignKey:referred_by;"`
	ProjectMembers        []ProjectMember
	Organizations         []Organization `gorm:"many2many:employee_organizations;"`
	Roles                 []Role         `gorm:"many2many:employee_roles;"`
	Positions             []Position     `gorm:"many2many:employee_positions;"`
	EmployeeRoles         []EmployeeRole
	EmployeePositions     []EmployeePosition
	EmployeeOrganizations []EmployeeOrganization
	EmployeeStacks        []EmployeeStack
	EmployeeChapters      []EmployeeChapter
	SocialAccounts        []SocialAccount
	Mentees               []*Employee `gorm:"references:ID;foreignKey:LineManagerID"`
	WorkUnitMembers       []WorkUnitMember
	Heads                 []ProjectHead

	EmployeeSalary *EmployeeSalary `json:"employee_salary"`
	BaseSalary     BaseSalary      `json:"base_salary"`

	DiscordAccount *DiscordAccount
}

// ToEmployeeMap create map from employees
func ToEmployeeMap(employees []*Employee) map[UUID]Employee {
	rs := map[UUID]Employee{}
	for _, e := range employees {
		rs[e.ID] = *e
	}

	return rs
}

func (e Employee) GetFirstNameFromFullName() string {
	if e.FullName == "" {
		return ""
	}
	s := strings.Split(e.FullName, " ")
	return s[len(s)-1]
}

type Employees []Employee

func (e Employees) ToTeamEmailIDMap() map[string]UUID {
	rs := map[string]UUID{}
	for _, e := range e {
		rs[e.TeamEmail] = e.ID
	}

	return rs
}

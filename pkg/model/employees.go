package model

import "time"

type WorkingStatus string

const (
	WorkingStatusPartTime  WorkingStatus = "part-time"
	WorkingStatusLeft      WorkingStatus = "left"
	WorkingStatusProbation WorkingStatus = "probation"
	WorkingStatusFullTime  WorkingStatus = "full-time"
)

func (e WorkingStatus) IsValid() bool {
	switch e {
	case
		WorkingStatusPartTime,
		WorkingStatusLeft,
		WorkingStatusProbation,
		WorkingStatusFullTime:
		return true
	}
	return false
}

type AccountStatus string

const (
	AccountStatusOnBoarding AccountStatus = "on-boarding"
	AccountStatusActive     AccountStatus = "active"
	AccountStatusProbation  AccountStatus = "probation"
	AccountStatusOnLeave    AccountStatus = "on-leave"
)

func (e AccountStatus) IsValid() bool {
	switch e {
	case
		AccountStatusOnBoarding,
		AccountStatusActive,
		AccountStatusProbation,
		AccountStatusOnLeave:
		return true
	}
	return false
}

type AccountRole string

const (
	AccountRoleAdmin  AccountRole = "admin"
	AccountRoleMember AccountRole = "member"
)

func (e AccountRole) IsValid() bool {
	switch e {
	case
		AccountRoleAdmin,
		AccountRoleMember:
		return true
	}
	return false
}

type Employee struct {
	BaseModel

	// basic info
	FullName               string
	DisplayName            string
	TeamEmail              string
	PersonalEmail          string
	Avatar                 string
	PhoneNumber            string
	Address                string
	MBTI                   string
	Gender                 string
	Horoscope              string
	PassportPhotoFront     string
	PassportPhotoBack      string
	IdentityCardPhotoFront string
	IdentityCardPhotoBack  string
	DateOfBirth            *time.Time

	// working info
	WorkingStatus WorkingStatus
	JoinedDate    *time.Time
	LeftDate      *time.Time
	ChapterID     UUID
	SeniorityID   UUID
	LineManagerID UUID
	AccountStatus AccountStatus

	// social services
	BasecampID             string
	BasecampAttachableSGID string `gorm:"column:basecamp_attachable_sgid"`
	GitlabID               string
	GithubID               string
	DiscordID              string
	NotionID               string

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

	Chapter        *Chapter
	Seniority      *Seniority
	LineManager    *Employee
	ProjectMembers []ProjectMember
	Roles          []Role     `gorm:"many2many:employee_roles;"`
	Positions      []Position `gorm:"many2many:employee_positions;"`
}

func (e AccountStatus) Valid() bool {
	switch e {
	case
		AccountStatusOnBoarding,
		AccountStatusActive,
		AccountStatusProbation,
		AccountStatusOnLeave:
		return true
	}
	return false
}

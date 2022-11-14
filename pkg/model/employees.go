package model

import "time"

type WorkingStatus string

const (
	WorkingStatusPartTime  WorkingStatus = "Part-time"
	WorkingStatusLeft      WorkingStatus = "Left"
	WorkingStatusProbation WorkingStatus = "Probation"
	WorkingStatusFullTime  WorkingStatus = "Full-time"
)

type AccountStatus string

const (
	AccountStatusOnBoarding AccountStatus = "on-boarding"
	AccountStatusActive     AccountStatus = "active"
	AccountStatusProbation  AccountStatus = "probation"
	AccountStatusOnLeave    AccountStatus = "on-leave"
)

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

	Chapter           *Chapter
	Seniority         *Seniority
	LineManager       *Employee
	EmployeePositions []EmployeePosition
	ProjectMembers    []ProjectMember
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

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
	FullName               string     `gorm:"column:full_name"`
	DisplayName            string     `gorm:"column:display_name"`
	TeamEmail              string     `gorm:"column:team_email"`
	PersonalEmail          string     `gorm:"column:personal_email"`
	Avatar                 string     `gorm:"column:avatar"`
	PhoneNumber            string     `gorm:"column:phone_number"`
	Address                string     `gorm:"column:address"`
	MBTI                   string     `gorm:"column:mbti"`
	Gender                 string     `gorm:"column:gender"`
	Horoscope              string     `gorm:"column:horoscope"`
	PassportPhotoFront     string     `gorm:"column:passport_photo_front"`
	PassportPhotoBack      string     `gorm:"column:passport_photo_back"`
	IdentityCardPhotoFront string     `gorm:"column:identity_card_photo_front"`
	IdentityCardPhotoBack  string     `gorm:"column:identity_card_photo_back"`
	DateOfBirth            *time.Time `gorm:"column:date_of_birth"`

	// working info
	WorkingStatus WorkingStatus `gorm:"column:working-status"`
	JoinedDate    *time.Time    `gorm:"column:joined_date"`
	LeftDate      *time.Time    `gorm:"column:left_date"`
	ChapterID     UUID          `gorm:"column:chapter_id"`
	SeniorityID   UUID          `gorm:"column:seniority_id"`
	PositionID    UUID          `gorm:"column:position_id"`
	LineManagerID UUID          `gorm:"column:line_manager_id"`
	AccountStatus AccountStatus `gorm:"column:account_status"`

	// social services
	BasecampID             string `gorm:"column:basecamp_id"`
	BasecampAttachableSGID string `gorm:"column:basecamp_attachable_sgid"`
	GitlabID               string `gorm:"column:gitlab_id"`
	GithubID               string `gorm:"column:github_id"`
	DiscordID              string `gorm:"column:discord_id"`

	// payroll info
	WiseRecipientEmail string `gorm:"column:wise_recipient_email"`
	WiseRecipientID    string `gorm:"column:wise_recipient_id"`
	WiseRecipientName  string `gorm:"column:wise_recipient_name"`
	WiseAccountNumber  string `gorm:"column:wise_account_number"`
	WiseCurrency       string `gorm:"column:wise_currency"`

	LocalBankBranch        string `gorm:"column:local_bank_branch"`
	LocalBankNumber        string `gorm:"column:local_bank_number"`
	LocalBankCurrency      string `gorm:"column:local_bank_currency"`
	LocalBranchName        string `gorm:"column:local_branch_name"`
	LocalBankRecipientName string `gorm:"column:local_bank_recipient_name"`

	Chapter       *Chapter       `gorm:"foreignkey:chapter_id"`
	Seniority     *Seniority     `gorm:"foreignkey:seniority_id"`
	Position      *Position      `gorm:"foreignkey:position_id"`
	LineManager   *Employee      `gorm:"foreignkey:line_manager_id"`
	EmployeeRoles *EmployeeRoles `gorm:"foreignkey:employee_id"`
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

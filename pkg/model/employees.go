package model

import "time"

type EmploymentStatus string

const (
	EmploymentStatusPartTime  EmploymentStatus = "Part-time"
	EmploymentStatusLeft      EmploymentStatus = "Left"
	EmploymentStatusProbation EmploymentStatus = "Probation"
	EmploymentStatusFullTime  EmploymentStatus = "Full-time"
)

type Employee struct {
	BaseModel

	// basic info
	FullName                string
	DisplayName             string
	TeamEmail               string
	PersonalEmail           string
	Avatar                  string
	PhoneNumber             string
	Address                 string
	MBTI                    string
	Gender                  string
	Horoscope               string
	PassportPhotoFront      string
	PassportPhotoBehind     string
	IdentityCardPhotoFront  string
	IdentityCardPhotoBehind string
	DateOfBirth             *time.Time

	// working info
	EmploymentStatus EmploymentStatus
	JoinedDate       *time.Time
	LeftDate         *time.Time

	// social services
	BasecampID             string
	BasecampAttachableSGID string
	GitlabID               string
	GithubID               string
	DiscordID              string

	// payroll info
	WiseRecipientEmail string
	WiseRecipientID    string
	WiseRecipientName  string
	WiseAccountNumber  string
	WiseCurrency       string

	LocalBankBranch        string
	LocalBankNumber        string
	LocalBankCurrency      string
	LocalBankName          string
	LocalBankRecipientName string
}

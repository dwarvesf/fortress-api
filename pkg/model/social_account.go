package model

type SocialAccount struct {
	BaseModel

	EmployeeID  UUID
	Type        SocialType
	AccountID   string
	Email       string
	DisplayName string
}

// SocialType social type for social_account table
type SocialType string

// values for working_status
const (
	SocialTypeGitHub   SocialType = "github"
	SocialTypeGitLab   SocialType = "gitlab"
	SocialTypeNotion   SocialType = "notion"
	SocialTypeDiscord  SocialType = "discord"
	SocialTypeLinkedIn SocialType = "linkedin"
	SocialTypeTwitter  SocialType = "twitter"
)

// IsValid validation for SocialType
func (e SocialType) IsValid() bool {
	switch e {
	case
		SocialTypeGitHub,
		SocialTypeGitLab,
		SocialTypeNotion,
		SocialTypeDiscord,
		SocialTypeLinkedIn,
		SocialTypeTwitter:
		return true
	}
	return false
}

// String returns the string type from the SocialType type
func (e SocialType) String() string {
	return string(e)
}

package model

type SocialAccount struct {
	BaseModel

	EmployeeID UUID
	Type       SocialAccountType
	AccountID  string
	Email      string
	Name       string
}

// SocialAccountType social type for social_account table
type SocialAccountType string

// values for working_status
const (
	SocialAccountTypeGitHub   SocialAccountType = "github"
	SocialAccountTypeGitLab   SocialAccountType = "gitlab"
	SocialAccountTypeNotion   SocialAccountType = "notion"
	SocialAccountTypeLinkedIn SocialAccountType = "linkedin"
	SocialAccountTypeTwitter  SocialAccountType = "twitter"
)

// IsValid validation for SocialAccountType
func (e SocialAccountType) IsValid() bool {
	switch e {
	case
		SocialAccountTypeGitHub,
		SocialAccountTypeGitLab,
		SocialAccountTypeNotion,
		SocialAccountTypeLinkedIn,
		SocialAccountTypeTwitter:
		return true
	}
	return false
}

// String returns the string type from the SocialAccountType type
func (e SocialAccountType) String() string {
	return string(e)
}

type SocialAccounts []SocialAccount

func (e SocialAccounts) GetGithub() *SocialAccount {
	for _, account := range e {
		if account.Type == SocialAccountTypeGitHub {
			return &account
		}
	}
	return nil
}

type SocialAccountInput struct {
	GithubID     string
	NotionID     string
	NotionName   string
	NotionEmail  string
	LinkedInName string
}

package employee

import "time"

type SearchFilter struct {
	WorkingStatus string
}

type EditGeneralInfo struct {
	Fullname      string
	Email         string
	Phone         string
	LineManagerID string
	DiscordID     string
	GithubID      string
	NotionID      string
}

type EditPersonalInfo struct {
	DoB           *time.Time
	Gender        string
	Address       string
	PersonalEmail string
}

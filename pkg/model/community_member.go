package model

import "github.com/lib/pq"

// CommunityMember represents information of a community member
type CommunityMember struct {
	BaseModel

	DiscordID       string         `json:"discord_id"`
	DiscordUsername string         `json:"discord_username"`
	Roles           pq.StringArray `json:"role" gorm:"type:text[]"`
	EmployeeID      *UUID          `json:"employee_id"`
	MemoUsername    string         `json:"memo_username"`
	GithubUsername  string         `json:"github_username"`
	PersonalEmail   string         `json:"personal_email"`
}

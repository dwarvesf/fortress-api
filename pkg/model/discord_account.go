package model

import "github.com/lib/pq"

type DiscordAccount struct {
	BaseModel

	DiscordID       string
	DiscordUsername string         `json:"discord_username"`
	MemoUsername    string         `json:"memo_username"`
	Roles           pq.StringArray `json:"role" gorm:"type:text[]"`
	GithubUsername  string         `json:"github_username"`
	PersonalEmail   string         `json:"personal_email"`

	Employee *Employee `json:"employee"`
}

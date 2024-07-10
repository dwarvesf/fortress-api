package model

import (
	"time"
)

type MemoPullRequest struct {
	Number         int       `json:"number"`
	Title          string    `json:"title"`
	DiscordId      string    `json:"discord_id"`
	GithubUserName string    `json:"github_user_name"`
	Url            string    `json:"url"`
	Timestamp      time.Time `json:"timestamp"`
}

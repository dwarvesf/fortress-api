package model

import (
	"time"
)

type MemoPullRequest struct {
	Number         int
	Title          string
	DiscordId      string
	GithubUserName string
	Url            string
	Timestamp      time.Time
}

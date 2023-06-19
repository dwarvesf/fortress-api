package model

import (
	"time"
)

type BraineryLog struct {
	BaseModel

	Title         string
	URL           string
	GithubID      string
	DiscordID     string
	EmployeeID    UUID
	Tags          JSONArrayString
	PublishedDate *time.Time
	Reward        float64
}

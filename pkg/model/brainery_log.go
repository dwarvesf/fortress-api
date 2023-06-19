package model

import (
	"time"

	"github.com/shopspring/decimal"
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
	Reward        decimal.Decimal
}

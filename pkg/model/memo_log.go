package model

import (
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type MemoLog struct {
	BaseModel

	Title       string
	URL         string
	Tags        JSONArrayString
	Description string
	PublishedAt *time.Time
	Reward      decimal.Decimal
	Category    pq.StringArray `json:"value" gorm:"type:text[]"`

	DiscordAccountIDs JSONArrayString `json:"discord_account_ids" gorm:"type:jsonb;column:discord_account_ids"`

	// This field is used to make sure response always contains authors
	AuthorMemoUsernames []string `json:"-" gorm:"-"`
}

type DiscordAccountMemoRank struct {
	DiscordID       string
	DiscordUsername string
	MemoUsername    string
	TotalMemos      int
	Rank            int
}

// Remove BeforeCreate method as we no longer use many-to-many join table

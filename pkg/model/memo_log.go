package model

import (
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
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

	Authors []DiscordAccount `json:"authors" gorm:"many2many:memo_authors;"`

	// This field is used to make sure response always contains authors
	AuthorMemoUsernames []string `json:"-" gorm:"-"`
}

func (MemoLog) BeforeCreate(db *gorm.DB) error {
	return db.SetupJoinTable(&MemoLog{}, "Authors", &MemoAuthor{})
}

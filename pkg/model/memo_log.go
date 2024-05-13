package model

import (
	"time"

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

	Authors []CommunityMember `json:"authors" gorm:"many2many:memo_authors;"`
}

func (MemoLog) BeforeCreate(db *gorm.DB) error {
	return db.SetupJoinTable(&MemoLog{}, "Authors", &MemoAuthor{})
}

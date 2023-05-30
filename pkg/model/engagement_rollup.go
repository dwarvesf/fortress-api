package model

import "github.com/shopspring/decimal"

type EngagementsRollup struct {
	BaseModel

	DiscordUserID   decimal.Decimal `gorm:"default:null"`
	LastMessageID   decimal.Decimal `gorm:"default:null"`
	DiscordUsername string          `gorm:"default:null"`
	ChannelID       decimal.Decimal `gorm:"default:null"`
	CategoryID      decimal.Decimal `gorm:"default:null"`
	MessageCount    int             `gorm:"default:null"`
	ReactionCount   int             `gorm:"default:null"`
}

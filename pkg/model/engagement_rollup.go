package model

type EngagementsRollup struct {
	BaseModel

	DiscordUserID   int64  `gorm:"default:null"`
	LastMessageID   int64  `gorm:"default:null"`
	DiscordUsername string `gorm:"default:null"`
	ChannelID       int64  `gorm:"default:null"`
	CategoryID      int64  `gorm:"default:null"`
	MessageCount    int    `gorm:"default:null"`
	ReactionCount   int    `gorm:"default:null"`
}

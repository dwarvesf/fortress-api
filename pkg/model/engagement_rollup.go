package model

type EngagementsRollup struct {
	BaseModel

	DiscordUserID   string `gorm:"default:null"`
	LatestMessageID string `gorm:"default:null"`
	DiscordUsername string `gorm:"default:null"`
	ChannelID       string `gorm:"default:null"`
	CategoryID      string `gorm:"default:null"`
	MessageCount    int    `gorm:"default:null"`
	ReactionCount   int    `gorm:"default:null"`
}

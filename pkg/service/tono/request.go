package tono

type GetGuildProfileResponse struct {
	Data GuildProfile `json:"data"`
}

type GuildProfile struct {
	ID           string         `json:"id"`
	CurrentLevel *ConfigXpLevel `json:"current_level"`
	NextLevel    *ConfigXpLevel `json:"next_level"`
	GuildXP      int            `json:"guild_xp"`
	NrOfActions  int            `json:"nr_of_actions"`
	Progress     float64        `json:"progress"`
	GuildRank    int            `json:"guild_rank"`
}

type ConfigXpLevel struct {
	Level int `json:"level" gorm:"primaryKey"`
	MinXP int `json:"min_xp"`
}

package engagement

type UpsertRollupRequest struct {
	DiscordUserID   string `json:"discordUserID" binding:"required"`
	LatestMessageID string `json:"latestMessageID" binding:"required"`
	ChannelID       string `json:"channelID" binding:"required"`
	CategoryID      string `json:"categoryID" binding:"required"`
	MessageCount    int    `json:"messageCount" binding:"required"`
	ReactionCount   int    `json:"reactionCount" binding:"required"`
}

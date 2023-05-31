package request

type UpsertRollupRequest struct {
	DiscordUserID string `json:"discordUserID" binding:"required"`
	LastMessageID string `json:"lastMessageID" binding:"required"`
	ChannelID     string `json:"channelID" binding:"required"`
	CategoryID    string `json:"categoryID" binding:"required"`
	MessageCount  int    `json:"messageCount" binding:"required"`
	ReactionCount int    `json:"reactionCount" binding:"required"`
}

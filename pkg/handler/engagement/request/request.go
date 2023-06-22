package request

import "errors"

var ErrInvalidCount = errors.New("message count or reaction count should be >0")

type UpsertRollupRequest struct {
	DiscordUserID string `json:"discordUserID" binding:"required"`
	LastMessageID string `json:"lastMessageID" binding:"required"`
	ChannelID     string `json:"channelID" binding:"required"`
	CategoryID    string `json:"categoryID"`
	MessageCount  int    `json:"messageCount"`
	ReactionCount int    `json:"reactionCount"`
}

func (r UpsertRollupRequest) Validate() error {
	if r.MessageCount == 0 && r.ReactionCount == 0 {
		return ErrInvalidCount
	}
	return nil
}

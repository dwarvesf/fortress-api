package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type DiscordTextMessageLog struct {
	ID         string    `json:"id"`
	Content    string    `json:"content"`
	AuthorName string    `json:"author_name"`
	AuthorID   string    `json:"author_id"`
	ChannelID  string    `json:"channel_id"`
	GuildID    string    `json:"guild_id"`
	Timestamp  time.Time `json:"timestamp"`
}

type ListDiscordTextMessageLog struct {
	Data []DiscordTextMessageLog `json:"data"`
} // @name ListDiscordTextMessageLog

func ToDiscordTextMessageLog(message model.DiscordTextMessageLog) DiscordTextMessageLog {
	return DiscordTextMessageLog{
		ID:         message.ID,
		Content:    message.Content,
		AuthorName: message.AuthorName,
		AuthorID:   message.AuthorID,
		ChannelID:  message.ChannelID,
		GuildID:    message.GuildID,
		Timestamp:  message.Timestamp,
	}
}

func ToListDiscordTextMessageLog(messages []model.DiscordTextMessageLog) []DiscordTextMessageLog {
	var results = make([]DiscordTextMessageLog, 0, len(messages))

	for _, message := range messages {
		results = append(results, ToDiscordTextMessageLog(message))
	}

	return results
}

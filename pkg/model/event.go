package model

import (
	"time"
)

// Event struct
type Event struct {
	BaseModel
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	Date             time.Time      `json:"date"`
	DiscordEventID   string         `json:"discord_event_id"`
	DiscordChannelID string         `json:"discord_channel_id"`
	DiscordMessageID string         `json:"discord_message_id"`
	DiscordCreatorID string         `json:"discord_creator_id"`
	EventType        EventType      `json:"type"`
	EventSpeakers    []EventSpeaker `json:"event_speakers"`
}

// EventSpeaker struct
type EventSpeaker struct {
	EventID          UUID   `json:"event_id"`
	DiscordAccountID UUID   `json:"discord_account_id"`
	Topic            string `json:"topic,omitempty"`
}

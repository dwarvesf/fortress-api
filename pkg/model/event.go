package model

import (
	"time"
)

// Event struct
type Event struct {
	BaseModel
	Name             string                    `json:"name"`
	Description      string                    `json:"description"`
	Date             time.Time                 `json:"date"`
	DiscordEventID   string                    `json:"discord_event_id"`
	DiscordChannelID string                    `json:"discord_channel_id"`
	DiscordMessageID string                    `json:"discord_message_id"`
	DiscordCreatorID string                    `json:"discord_creator_id"`
	EventType        DiscordScheduledEventType `json:"type"`
	EventSpeakers    []EventSpeaker            `json:"event_speakers"`
	IsOver           bool                      `json:"is_over" gorm:"-"`
}

// EventSpeaker struct
type EventSpeaker struct {
	EventID          UUID   `json:"event_id"`
	DiscordAccountID UUID   `json:"discord_account_id"`
	Topic            string `json:"topic,omitempty"`
}

type DiscordScheduledEventType string

const (
	DiscordScheduledEventTypeOGIF DiscordScheduledEventType = "OGIF"
	DiscordScheduledEventTypeDemo DiscordScheduledEventType = "DEMO"
)

// IsValid validation for DiscordScheduledEventType
func (e DiscordScheduledEventType) IsValid() bool {
	switch e {
	case
		DiscordScheduledEventTypeOGIF,
		DiscordScheduledEventTypeDemo:
		return true
	}
	return false
}

// String returns the string type from the DiscordScheduledEventType type
func (e DiscordScheduledEventType) String() string {
	return string(e)
}

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
	Image            string                    `json:"image" gorm:"-"`
	DiscordEventID   string                    `json:"discordEventId"`
	DiscordChannelID string                    `json:"discordChannelId"`
	DiscordMessageID string                    `json:"discordMessageId"`
	DiscordCreatorID string                    `json:"discordCreatorId"`
	EventType        DiscordScheduledEventType `json:"type"`
	EventSpeakers    []EventSpeaker            `json:"eventSpeakers"`
	IsOver           bool                      `json:"isOver" gorm:"-"`
}

// EventSpeaker struct
type EventSpeaker struct {
	EventID          UUID   `json:"eventId"`
	DiscordAccountID UUID   `json:"discordAccountId"`
	Topic            string `json:"topic,omitempty"`

	Event *Event `json:"event"`
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

// OgifLeaderboardRecord represents an element in the OGIF leaderboard
type OgifLeaderboardRecord struct {
	DiscordID  string `json:"discordID"`
	SpeakCount int64  `json:"speakCount"`
}

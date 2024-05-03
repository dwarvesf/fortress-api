package model

import (
	"time"
)

type DiscordEventType string

const (
	DiscordEventTypeOGIF DiscordEventType = "OGIF"
	DiscordEventTypeDemo DiscordEventType = "DEMO"
)

func (i DiscordEventType) IsValid() bool {
	switch i {
	case DiscordEventTypeOGIF,
		DiscordEventTypeDemo:
		return true
	}
	return false
}

func (i DiscordEventType) String() string {
	return string(i)
}

// DiscordEvent struct
type DiscordEvent struct {
	BaseModel
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	Date             time.Time        `json:"date"`
	EventURL         string           `json:"event_url"`
	MsgURL           string           `json:"msg_url"`
	DiscordEventType DiscordEventType `json:"type"`
	EventSpeakers    []EventSpeaker   `json:"event_speakers"`
}

// EventSpeaker struct
type EventSpeaker struct {
	DiscordEventID   UUID   `json:"discord_event_id"`
	DiscordAccountID UUID   `json:"discord_account_id"`
	Topic            string `json:"topic,omitempty"`
}

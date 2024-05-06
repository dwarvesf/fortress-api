package model

import (
	"time"
)

// Event struct
type Event struct {
	BaseModel
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Date          time.Time      `json:"date"`
	EventURL      string         `json:"event_url"`
	MsgURL        string         `json:"msg_url"`
	EventType     EventType      `json:"type"`
	EventSpeakers []EventSpeaker `json:"event_speakers"`
}

// EventSpeaker struct
type EventSpeaker struct {
	EventID          UUID   `json:"event_id"`
	DiscordAccountID UUID   `json:"discord_account_id"`
	Topic            string `json:"topic,omitempty"`
}

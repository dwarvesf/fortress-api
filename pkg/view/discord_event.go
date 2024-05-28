package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// Event struct
type DiscordEvent struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Date             time.Time  `json:"date"`
	DiscordEventID   string     `json:"discord_event_id"`
	DiscordChannelID string     `json:"discord_channel_id"`
	DiscordMessageID string     `json:"discord_message_id"`
	DiscordCreatorID string     `json:"discord_creator_id"`
	EventType        string     `json:"type"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        *time.Time `json:"updatedAt"`
}

func ToDiscordEvent(event model.Event) DiscordEvent {
	return DiscordEvent{
		ID:               event.ID.String(),
		Name:             event.Name,
		Description:      event.Description,
		Date:             event.Date,
		EventType:        event.EventType.String(),
		DiscordEventID:   event.DiscordEventID,
		DiscordChannelID: event.DiscordChannelID,
		DiscordMessageID: event.DiscordMessageID,
		DiscordCreatorID: event.DiscordCreatorID,
		CreatedAt:        event.CreatedAt,
		UpdatedAt:        event.UpdatedAt,
	}
}

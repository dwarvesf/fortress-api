package request

import (
	"errors"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/handler/discord/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type BraineryReportInput struct {
	View      string `json:"view" binding:"required"`
	ChannelID string `json:"channelID" binding:"required"`
}

func (input BraineryReportInput) Validate() error {
	if len(input.View) == 0 {
		return errs.ErrEmptyReportView
	}

	if len(input.ChannelID) == 0 {
		return errs.ErrEmptyChannelID
	}
	return nil
}

type DeliveryMetricReportInput struct {
	View               string `json:"view" binding:"required"`
	ChannelID          string `json:"channelID" binding:"required"`
	OnlyCompletedMonth bool   `json:"onlyCompletedMonth"`
	Sync               bool   `json:"sync"`
}

func (input DeliveryMetricReportInput) Validate() error {
	if len(input.View) == 0 {
		return errs.ErrEmptyReportView
	}

	if len(input.ChannelID) == 0 {
		return errs.ErrEmptyChannelID
	}
	return nil
}

type DiscordEventInput struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Date             time.Time `json:"date"`
	Image            string    `json:"image"`
	DiscordEventID   string    `json:"discord_event_id"`
	DiscordChannelID string    `json:"discord_channel_id"`
	DiscordCreatorID string    `json:"discord_creator_id"`
	DiscordMessageID string    `json:"discord_message_id"`
}

func (input DiscordEventInput) Validate() error {
	if len(input.Name) == 0 {
		return errs.ErrEmptyName
	}
	if len(input.DiscordChannelID) == 0 {
		return errs.ErrEmptyChannelID
	}
	if len(input.DiscordCreatorID) == 0 {
		return errs.ErrEmptyCreatorID
	}
	return nil
}

func (in DiscordEventInput) EventType() (model.DiscordScheduledEventType, error) {
	switch {
	case strings.Contains(strings.ToLower(in.Description), "demo"),
		strings.Contains(strings.ToLower(in.Name), "demo"),
		strings.Contains(strings.ToLower(in.Description), "showcase"),
		strings.Contains(strings.ToLower(in.Name), "showcase"):
		return model.DiscordScheduledEventTypeDemo, nil
	case strings.Contains(strings.ToLower(in.Description), "ogif"),
		strings.Contains(strings.ToLower(in.Name), "ogif"):
		return model.DiscordScheduledEventTypeOGIF, nil
	default:
		return model.DiscordScheduledEventType(""), errors.New("invalid event type")
	}
}

type DiscordEventSpeakerInput struct {
	ID    string `json:"id"`
	Topic string `json:"topic"`
}

func (input DiscordEventSpeakerInput) Validate() error {
	if len(input.ID) == 0 {
		return errs.ErrEmptyID
	}
	if len(input.Topic) == 0 {
		return errs.ErrEmptyTopic
	}
	return nil
}

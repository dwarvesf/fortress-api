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
	ID          string          `json:"id"`
	Name        string          `json:"title"`
	Description string          `json:"description"`
	Date        *time.Time      `json:"date"`
	ChannelID   string          `json:"channel_id"`
	CreatorID   string          `json:"creator_id"`
	MessageID   string          `json:"message_id"`
	EventType   model.EventType `json:"event_type"`
}

func (input DiscordEventInput) Validate() error {
	if len(input.Name) == 0 {
		return errs.ErrEmptyName
	}
	if len(input.ChannelID) == 0 {
		return errs.ErrEmptyChannelID
	}
	if len(input.CreatorID) == 0 {
		return errs.ErrEmptyCreatorID
	}
	if input.Date == nil {
		return errs.ErrEmptyDate
	}
	return nil
}

func (in *DiscordEventInput) SetEventType() error {
	switch {
	case (strings.Contains(strings.ToLower(in.Description), "demo") || strings.Contains(strings.ToLower(in.Name), "demo")) || (strings.Contains(strings.ToLower(in.Description), "showcase") || strings.Contains(strings.ToLower(in.Name), "showcase")):
		in.EventType = model.EventTypeDiscordDemo
	case strings.Contains(strings.ToLower(in.Description), "ogif") || strings.Contains(strings.ToLower(in.Name), "ogif"):
		in.EventType = model.EventTypeDiscordOGIF
	default:
		return errors.New("event type not found")
	}
	return nil
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

package discord

import (
	"github.com/bwmarrin/discordgo"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IService interface {
	PostBirthdayMsg(msg string) (model.DiscordMessage, error)
	GetMembers() ([]*discordgo.Member, error)

	// CreateEvent create a discord event
	CreateEvent(event *model.Schedule) (scheduledEvent *discordgo.GuildScheduledEvent, err error)
	UpdateEvent(event *model.Schedule) (scheduledEvent *discordgo.GuildScheduledEvent, err error)
	DeleteEvent(event *model.Schedule) error

	// Log discord
	SendMessage(msg, webhookUrl string) (*model.DiscordMessage, error)
}

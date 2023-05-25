package discord

import (
	"github.com/bwmarrin/discordgo"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IService interface {
	PostBirthdayMsg(msg string) (model.DiscordMessage, error)
	GetMembers() ([]*discordgo.Member, error)
	GetMember(userID string) (*discordgo.Member, error)
	SearchMember(discordName string) ([]*discordgo.Member, error)
	GetRoles() (Roles, error)
	AddRole(userID, roleID string) error
	RemoveRole(userID string, roleID string) error

	// CreateEvent create a discord event
	CreateEvent(event *model.Schedule) (scheduledEvent *discordgo.GuildScheduledEvent, err error)
	UpdateEvent(event *model.Schedule) (scheduledEvent *discordgo.GuildScheduledEvent, err error)
	DeleteEvent(event *model.Schedule) error

	// SendMessage logs a message to a Discord channel
	SendMessage(msg, webhookUrl string) (*model.DiscordMessage, error)
}

package discord

import (
	"github.com/bwmarrin/discordgo"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type IService interface {
	PostBirthdayMsg(msg string) (model.DiscordMessage, error)
	GetMembers() ([]*discordgo.Member, error)
	GetMember(userID string) (*discordgo.Member, error)
	GetMemberByUsername(username string) (*discordgo.Member, error)
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

	GetChannels() ([]*discordgo.Channel, error)
	GetMessagesAfterCursor(channelID string, cursorMessageID string, lastMessageID string) ([]*discordgo.Message, error)
	ReportBraineryMetrics(queryView string, braineryMetric *view.BraineryMetric, channelID string) (*discordgo.Message, error)
	SendEmbeddedMessageWithChannel(original *model.OriginalDiscordMessage, embed *discordgo.MessageEmbed, channelId string) (*discordgo.Message, error)
	DeliveryMetricWeeklyReport(deliveryMetrics *view.DeliveryMetricWeeklyReport, channelID string) (*discordgo.Message, error)
}

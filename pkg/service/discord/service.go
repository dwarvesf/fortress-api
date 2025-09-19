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

	CreateEvent(event *model.Schedule) (scheduledEvent *discordgo.GuildScheduledEvent, err error)
	UpdateEvent(event *model.Schedule) (scheduledEvent *discordgo.GuildScheduledEvent, err error)
	DeleteEvent(event *model.Schedule) error
	ListEvents() ([]*discordgo.GuildScheduledEvent, error)

	GetChannels() ([]*discordgo.Channel, error)
	GetMessagesAfterCursor(channelID string, cursorMessageID string, lastMessageID string) ([]*discordgo.Message, error)
	GetChannelMessages(channelID, before, after string, limit int) ([]*discordgo.Message, error)
	GetEventByID(eventID string) (*discordgo.GuildScheduledEvent, error)

	ReportBraineryMetrics(queryView string, braineryMetric *view.BraineryMetric, channelID string) (*discordgo.Message, error)
	DeliveryMetricWeeklyReport(deliveryMetrics *view.DeliveryMetricWeeklyReport, leaderBoard *view.WeeklyLeaderBoard, channelID string) (*discordgo.Message, error)
	DeliveryMetricMonthlyReport(deliveryMetrics *view.DeliveryMetricMonthlyReport, leaderBoard *view.WeeklyLeaderBoard, channelID string) (*discordgo.Message, error)
	SendNewMemoMessage(
		guildID string,
		memos []model.MemoLog,
		channelID string,
		getDiscordAccountByID func(discordAccountID string) (*model.DiscordAccount, error),
	) (*discordgo.Message, error)
	SendWeeklyMemosMessage(
		guildID string,
		memos []model.MemoLog,
		weekRangeStr,
		channelID string,
		getDiscordAccountByID func(discordAccountID string) (*model.DiscordAccount, error),
		newAuthors []string,
		resolveAuthorsByTitle func(title string) ([]string, error),
		getDiscordIDByUsername func(username string) (string, error),
	) (*discordgo.Message, error)
	SendMonthlyMemosMessage(
		guildID string,
		memos []model.MemoLog,
		monthRangeStr,
		channelID string,
		getDiscordAccountByID func(discordAccountID string) (*model.DiscordAccount, error),
		newAuthors []string,
		getDiscordIDByUsername func(username string) (string, error),
	) (*discordgo.Message, error)
	SendLeaderboardMessage(
		guildID string,
		period string, // "weekly" or "monthly"
		channelID string,
		getDiscordAccountByID func(discordAccountID string) (*model.DiscordAccount, error),
		getDiscordIDByUsername func(username string) (string, error),
		getAllTimeMemos func() ([]model.MemoLog, error), // Function to fetch all-time memos since July 2025
	) (*discordgo.Message, error)
	/*
		WEBHOOK
	*/

	// SendMessage logs a message to a Discord channel through a webhook
	SendMessage(discordMsg model.DiscordMessage, webhookUrl string) (*model.DiscordMessage, error)
	SendEmbeddedMessageWithChannel(original *model.OriginalDiscordMessage, embed *discordgo.MessageEmbed, channelId string) (*discordgo.Message, error)
	SendDiscordMessageWithChannel(ses *discordgo.Session, msg *discordgo.Message, channelId string) error

	ListActiveThreadsByChannelID(guildID, channelID string) ([]discordgo.Channel, error)
}

package model

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordMessage struct {
	AvatarURL  string                     `json:"avatar_url"`
	Content    string                     `json:"content"`
	Embeds     []DiscordMessageEmbed      `json:"embeds"`
	Components []DiscordMessageComponents `json:"components"`
}

type DiscordMessageEmbed struct {
	Author      DiscordMessageAuthor  `json:"author"`
	Title       string                `json:"title"`
	URL         string                `json:"url"`
	Description string                `json:"description"`
	Color       int64                 `json:"color"`
	Fields      []DiscordMessageField `json:"fields"`
	Thumbnail   DiscordMessageImage   `json:"thumbnail"`
	Image       DiscordMessageImage   `json:"image"`
	Footer      DiscordMessageFooter  `json:"footer"`
	Timestamp   string                `json:"timestamp"`
}

type DiscordMessageAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IconURL string `json:"icon_url"`
}

type DiscordMessageField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline *bool  `json:"inline,omitempty"`
}

type DiscordMessageFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

type DiscordMessageImage struct {
	URL string `json:"url"`
}

type DiscordMessageComponents struct {
	Components []DiscordMessageComponent `json:"components"`
	Type       int64                     `json:"type"`
}

type DiscordMessageComponent struct {
	CustomID string      `json:"custom_id"`
	Disabled bool        `json:"disabled"`
	Emoji    interface{} `json:"emoji"`
	Label    string      `json:"label"`
	Style    int64       `json:"style"`
	Type     int64       `json:"type"`
	URL      interface{} `json:"url"`
}

type OriginalDiscordMessage struct {
	RawContent  string
	ContentArgs []string
	ChannelId   string
	GuildId     string
	Author      *discordgo.User
	Roles       []string
}

type DiscordTextMessageLog struct {
	ID         string
	Content    string
	AuthorName string
	AuthorID   string
	ChannelID  string
	GuildID    string
	Timestamp  time.Time
}

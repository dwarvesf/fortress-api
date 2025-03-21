package discord

import (
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// normalize add some default to embedded message if not set
func normalize(original *model.OriginalDiscordMessage, response *discordgo.MessageEmbed) *discordgo.MessageEmbed {
	if response.Timestamp == "" {
		response.Timestamp = time.Now().Format(time.RFC3339)
	}

	// I did something tricky here, if timestamp is custom, we don't want to show it, because in case of user want to add a custom date time format in the footer
	// instead of automatically add it, we don't want to show it twice.
	if response.Timestamp == "custom" {
		response.Timestamp = ""
	}

	if response.Color == 0 {
		// default df color #D14960
		response.Color = 13715808
	}
	if response.Footer == nil {
		response.Footer = &discordgo.MessageEmbedFooter{
			IconURL: "https://cdn.discordapp.com/avatars/564764617545482251/9c9bd4aaba164fc0b92f13f052405b4d.webp?size=160",
			Text:    "?help to see all commands",
		}
	}
	return response
}

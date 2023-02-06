package discord

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type DiscordService interface {
	PostBirthdayMsg(msg string) (model.DiscordMessage, error)
}

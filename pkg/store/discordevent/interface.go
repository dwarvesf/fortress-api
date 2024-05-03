package discordevent

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, query *Query) (e *model.DiscordEvent, err error)
	All(db *gorm.DB, public bool, preload bool) ([]*model.DiscordEvent, error)
	Create(db *gorm.DB, e *model.DiscordEvent) (de *model.DiscordEvent, err error)
}

// Query present invoice query from user
type Query struct {
	ID                string
	EventURL          string
	MsgURL            string
	DiscordEventTypes []model.DiscordEventType
}

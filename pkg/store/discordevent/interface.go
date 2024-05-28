package discordevent

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, query *Query) (e *model.Event, err error)
	All(db *gorm.DB, query *Query, preload bool) ([]*model.Event, error)
	Create(db *gorm.DB, e *model.Event) (de *model.Event, err error)
	SetSpeakers(db *gorm.DB, e *model.Event) error
}

// Query present invoice query from user
type Query struct {
	ID             string
	DiscordEventID string
	EventTypes     []model.EventType
	Limit          int
	Offset         int
	After          *time.Time
}

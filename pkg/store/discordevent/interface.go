package discordevent

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, query *Query) (e *model.Event, err error)
	All(db *gorm.DB, public bool, preload bool) ([]*model.Event, error)
	Create(db *gorm.DB, e *model.Event) (de *model.Event, err error)
}

// Query present invoice query from user
type Query struct {
	ID         string
	EventURL   string
	MsgURL     string
	EventTypes []model.EventType
}

package discordevent

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get client by id
func (s *store) One(db *gorm.DB, query *Query) (*model.Event, error) {
	var e *model.Event
	if query.ID != "" {
		db = db.Where("id = ?", query.ID)
	}
	if query.DiscordEventID != "" {
		db = db.Where("discord_event_id = ?", query.DiscordEventID)
	}
	return e, db.Preload("EventSpeakers").First(&e).Error
}

// All get all client
func (s *store) All(db *gorm.DB, public bool, preload bool) ([]*model.Event, error) {
	var e []*model.Event

	query := db.Preload("EventSpeakers", "deleted_at IS NULL")

	return e, query.Find(&e).Error
}

// Create creates a new e
func (s *store) Create(db *gorm.DB, e *model.Event) (*model.Event, error) {
	return e, db.Create(e).Error
}

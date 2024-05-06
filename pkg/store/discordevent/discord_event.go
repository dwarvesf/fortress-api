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
	if query.EventURL != "" {
		db = db.Where("event_url = ?", query.EventURL)
	}
	if query.MsgURL != "" {
		db = db.Where("msg_url = ?", query.MsgURL)
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

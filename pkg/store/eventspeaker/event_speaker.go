package eventspeaker

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create creates a new e
func (s *store) Create(db *gorm.DB, e *model.EventSpeaker) (*model.EventSpeaker, error) {
	return e, db.Create(e).Error
}

// DeleteAllByEventID deletes all event speakers by event id
func (s *store) DeleteAllByEventID(db *gorm.DB, eventID string) error {
	return db.Table("event_speakers").Where("event_id = ?", eventID).Delete(&model.EventSpeaker{}).Error
}

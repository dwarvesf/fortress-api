package eventspeaker

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, s *model.EventSpeaker) (ep *model.EventSpeaker, err error)
	DeleteAllByEventID(db *gorm.DB, eventID string) error
}

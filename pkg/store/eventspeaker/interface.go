package eventspeaker

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, s *model.EventSpeaker) (ep *model.EventSpeaker, err error)
	DeleteAllByEventID(db *gorm.DB, eventID string) error
	List(db *gorm.DB, discordID string, after *time.Time, topic string) ([]model.EventSpeaker, error)
	GetSpeakerStats(db *gorm.DB, discordID string, after *time.Time, topic string) (SpeakerStats, error)
	Count(db *gorm.DB, discordID string, after *time.Time, topic string) (int64, error)
}

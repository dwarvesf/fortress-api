package feedbackevent

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	IsExist(db *gorm.DB, id string) (bool, error)
	One(db *gorm.DB, id string, preload bool) (event *model.FeedbackEvent, err error)
	OneByTypeInTimeRange(db *gorm.DB, eventType model.EventType, eventSubtype model.EventSubtype, from, to *time.Time) (*model.FeedbackEvent, error)
	GetBySubtype(db *gorm.DB, subtype string, pagination model.Pagination) (events []*model.FeedbackEvent, total int64, err error)
	Create(db *gorm.DB, feedbackEvent *model.FeedbackEvent) (*model.FeedbackEvent, error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.FeedbackEvent, updatedFields ...string) (event *model.FeedbackEvent, err error)
	DeleteByID(db *gorm.DB, id string) error
	GetLatestEventByType(db *gorm.DB, eventType model.EventType, eventSubtype model.EventSubtype, num int) ([]*model.FeedbackEvent, error)
}

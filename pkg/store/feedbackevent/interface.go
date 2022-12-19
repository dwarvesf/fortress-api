package feedbackevent

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	IsExist(db *gorm.DB, id string) (bool, error)
	GetBySubtypeAndProjectIDs(db *gorm.DB, subtype string, projectIDs []string, pagination model.Pagination) (events []*model.FeedbackEvent, total int64, err error)
	GetByTypeInTimeRange(db *gorm.DB, eventType model.EventType, eventSubtype model.EventSubtype, from, to *time.Time) (*model.FeedbackEvent, error)
	Create(db *gorm.DB, feedbackEvent *model.FeedbackEvent) (*model.FeedbackEvent, error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.FeedbackEvent, updatedFields ...string) (event *model.FeedbackEvent, err error)
	One(db *gorm.DB, id string) (event *model.FeedbackEvent, err error)
	DeleteByID(db *gorm.DB, id string) error
}

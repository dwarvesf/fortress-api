package feedbackevent

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// IsExist check the existence of FeedbackEvent
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM feedback_events WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}

// GetBySubtypeWithPagination return list of FeedbackEvent by subtype and pagination
func (s *store) GetBySubtypeWithPagination(db *gorm.DB, subtype string, pagination model.Pagination) ([]*model.FeedbackEvent, int64, error) {
	var events []*model.FeedbackEvent
	var total int64

	query := db.Table("feedback_events").
		Where("subtype = ?", subtype).
		Count(&total).
		Order(pagination.Sort)

	if pagination.Sort == "" {
		query = query.Order("created_at DESC")
	}

	limit, offset := pagination.ToLimitOffset()
	if pagination.Page > 0 {
		query = query.Limit(limit)
	}

	query = query.Offset(offset)

	return events, total, query.
		Preload("Employee", "deleted_at IS NULL").
		Preload("Topics", "deleted_at IS NULL").
		Preload("Topics.EmployeeEventReviewers", "deleted_at IS NULL").
		Find(&events).Error
}

// Create create new one
func (s *store) Create(db *gorm.DB, feedbackEvent *model.FeedbackEvent) (*model.FeedbackEvent, error) {
	return feedbackEvent, db.Create(&feedbackEvent).Error
}

// One get 1 by id
func (s *store) One(db *gorm.DB, id string) (*model.FeedbackEvent, error) {
	var event *model.FeedbackEvent
	return event, db.Where("id = ?", id).
		Preload("Employee", "deleted_at IS NULL").
		First(&event).Error
}

// One get 1 by id
func (s *store) GetByTypeInTimeRange(db *gorm.DB, eventType model.EventType, eventSubtype model.EventSubtype, from, to *time.Time) (*model.FeedbackEvent, error) {
	var event *model.FeedbackEvent
	return event, db.Where("type = ?", eventType).
		Where("subtype = ?", eventSubtype).
		Where("start_date = ? AND end_date = ?", from, to).
		First(&event).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.FeedbackEvent, updatedFields ...string) (*model.FeedbackEvent, error) {
	feedbackEvent := model.FeedbackEvent{}
	return &feedbackEvent, db.Model(&feedbackEvent).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

// DeleteByID delete FeedbackEvent by id
func (s *store) DeleteByID(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.FeedbackEvent{}).Error
}

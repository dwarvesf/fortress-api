package feedbackevent

import (
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

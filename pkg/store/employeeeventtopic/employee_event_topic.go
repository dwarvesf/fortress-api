package employeeeventtopic

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// GetByEmployeeIDWithPagination return list of EmployeeEventTopic by employeeID and pagination
func (s *store) GetByEmployeeIDWithPagination(db *gorm.DB, employeeID string, input GetByEmployeeIDInput, pagination model.Pagination) ([]*model.EmployeeEventTopic, int64, error) {
	var eTopics []*model.EmployeeEventTopic
	var total int64

	query := db.
		Table("employee_event_topics").
		Joins("JOIN employee_event_reviewers eer ON employee_event_topics.id = eer.employee_event_topic_id").
		Joins("JOIN feedback_events fe ON employee_event_topics.event_id = fe.id").
		Where("(eer.reviewer_id = ? OR (employee_event_topics.employee_id = ? AND fe.type = ?))",
			employeeID,
			employeeID,
			model.EventTypeFeedback)

	if input.Status != "" {
		query = query.Where("eer.status = ?", input.Status)
	}

	query = query.Count(&total)
	query = query.Order(pagination.Sort)

	limit, offset := pagination.ToLimitOffset()
	if pagination.Page > 0 {
		query = query.Limit(limit)
	}

	query = query.Offset(offset)

	return eTopics, total, query.
		Preload("Event", "deleted_at IS NULL").
		Preload("Event.Employee", "deleted_at IS NULL").
		Preload("EmployeeEventReviewers", "deleted_at IS NULL").
		Preload("EmployeeEventReviewers.Reviewer", "deleted_at IS NULL").
		Find(&eTopics).Error
}
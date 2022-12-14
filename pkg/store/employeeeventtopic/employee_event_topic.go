package employeeeventtopic

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get topic by id
func (s *store) One(db *gorm.DB, id string, eventID string) (*model.EmployeeEventTopic, error) {
	var topic *model.EmployeeEventTopic

	// TODO: check Preload
	return topic, db.Where("id = ? AND event_id = ?", id, eventID).
		Preload("Employee", "deleted_at IS NULL").
		Preload("Event", "deleted_at IS NULL").
		Preload("EmployeeEventReviewers", "deleted_at IS NULL").
		Preload("EmployeeEventReviewers.Reviewer", "deleted_at IS NULL").
		First(&topic).Error
}

// All get topic by id
func (s *store) All(db *gorm.DB, eventID string) ([]*model.EmployeeEventTopic, error) {
	var topics []*model.EmployeeEventTopic

	return topics, db.Where("event_id = ?", eventID).Find(&topics).Error
}

// GetByEmployeeIDWithPagination return list of EmployeeEventTopic by employeeID and pagination
func (s *store) GetByEmployeeIDWithPagination(db *gorm.DB, employeeID string, input GetByEmployeeIDInput, pagination model.Pagination) ([]*model.EmployeeEventTopic, int64, error) {
	var eTopics []*model.EmployeeEventTopic
	var total int64

	query := db.
		Table("employee_event_topics").
		Joins("JOIN employee_event_reviewers eer ON employee_event_topics.id = eer.employee_event_topic_id").
		Joins("JOIN feedback_events fe ON employee_event_topics.event_id = fe.id").
		Where("((eer.reviewer_id = ? AND fe.type = ?) OR (employee_event_topics.employee_id = ? AND fe.type = ?)) AND eer.reviewer_status <> ?",
			employeeID,
			model.EventTypeSurvey,
			employeeID,
			model.EventTypeFeedback,
			model.EventReviewerStatusNone,
		)

	if input.Status != "" {
		query = query.Where("eer.reviewer_status = ?", input.Status)
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
		Preload("EmployeeEventReviewers", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN employee_event_topics  ON employee_event_topics.id = employee_event_reviewers.employee_event_topic_id").
				Joins("JOIN feedback_events fe ON employee_event_topics.event_id = fe.id").
				Where("((employee_event_reviewers.reviewer_id = ? AND fe.type = ?) OR (employee_event_topics.employee_id = ? AND fe.type = ?)) AND employee_event_reviewers.reviewer_status <> ?",
					employeeID,
					model.EventTypeSurvey,
					employeeID,
					model.EventTypeFeedback,
					model.EventReviewerStatusNone,
				)
		}).
		Preload("EmployeeEventReviewers.Reviewer", "deleted_at IS NULL").
		Find(&eTopics).Error
}

// GetByEventIDWithPagination return list of EmployeeEventTopic by eventID and pagination
func (s *store) GetByEventIDWithPagination(db *gorm.DB, eventID string, pagination model.Pagination) ([]*model.EmployeeEventTopic, int64, error) {
	var topics []*model.EmployeeEventTopic
	var total int64

	query := db.
		Table("employee_event_topics").
		Where("event_id = ? AND deleted_at IS NULL", eventID).
		Count(&total).
		Order(pagination.Sort)

	limit, offset := pagination.ToLimitOffset()
	if pagination.Page > 0 {
		query = query.Limit(limit)
	}

	query = query.Offset(offset)

	return topics, total, query.
		Preload("Event", "deleted_at IS NULL").
		Preload("Employee", "deleted_at IS NULL").
		Preload("EmployeeEventReviewers", "deleted_at IS NULL").
		Preload("EmployeeEventReviewers.Reviewer", "deleted_at IS NULL").
		Find(&topics).Error
}

// BatchCreate create new one
func (s *store) BatchCreate(db *gorm.DB, employeeEventTopics []model.EmployeeEventTopic) ([]model.EmployeeEventTopic, error) {
	return employeeEventTopics, db.Create(&employeeEventTopics).Error
}

// Create create new one
func (s *store) DeleteByEventID(db *gorm.DB, eventID string) error {
	return db.Where("event_id = ?", eventID).Delete(&model.EmployeeEventTopic{}).Error
}

// DeleteByID
func (s *store) DeleteByID(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.EmployeeEventTopic{}).Error
}

package employeeeventreviewer

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get one record by reviewerID and topicID
func (s *store) One(db *gorm.DB, reviewerID string, topicID string) (*model.EmployeeEventReviewer, error) {
	var employeeeventreviewer *model.EmployeeEventReviewer
	return employeeeventreviewer, db.Where("reviewer_id = ? AND employee_event_topic_id = ? ", reviewerID, topicID).First(&employeeeventreviewer).Error
}

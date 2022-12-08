package employeeeventquestion

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetByEventReviewerID(db *gorm.DB, reviewID string) ([]*model.EmployeeEventQuestion, error) {
	var eventQuestions []*model.EmployeeEventQuestion
	return eventQuestions, db.Where("employee_event_reviewer_id = ?", reviewID).Order("\"order\"").Find(&eventQuestions).Error
}

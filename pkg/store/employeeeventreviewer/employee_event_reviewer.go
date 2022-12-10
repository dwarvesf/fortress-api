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

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.EmployeeEventReviewer, updatedFields ...string) (*model.EmployeeEventReviewer, error) {
	eventReviewer := model.EmployeeEventReviewer{}
	return &eventReviewer, db.Model(&eventReviewer).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

// Create create new one
func (s *store) BatchCreate(db *gorm.DB, employeeEventReviewers []model.EmployeeEventReviewer) ([]model.EmployeeEventReviewer, error) {
	return employeeEventReviewers, db.Create(&employeeEventReviewers).Error
}

// Create a employee event reviewer
func (s *store) Create(tx *gorm.DB, eventReviewer *model.EmployeeEventReviewer) (*model.EmployeeEventReviewer, error) {
	return eventReviewer, tx.Create(&eventReviewer).Error
}

// DeleteByEventID delete EmployeeEventReviewer by eventID
func (s *store) DeleteByEventID(db *gorm.DB, eventID string) error {
	return db.Where("event_id = ?", eventID).Delete(&model.EmployeeEventReviewer{}).Error
}

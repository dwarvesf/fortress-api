package employeeeventreviewer

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get one record by ID
func (s *store) One(db *gorm.DB, id string) (*model.EmployeeEventReviewer, error) {
	var eer *model.EmployeeEventReviewer
	return eer, db.Where("id = ?", id).
		Preload("Reviewer", "deleted_at IS NULL").
		First(&eer).Error
}

// GetByReviewerID get one record by reviewerID and topicID
func (s *store) GetByReviewerID(db *gorm.DB, reviewerID string, topicID string) (*model.EmployeeEventReviewer, error) {
	var eer *model.EmployeeEventReviewer
	return eer, db.Where("reviewer_id = ? AND employee_event_topic_id = ? ", reviewerID, topicID).First(&eer).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.EmployeeEventReviewer, updatedFields ...string) (*model.EmployeeEventReviewer, error) {
	eventReviewer := model.EmployeeEventReviewer{}
	return &eventReviewer, db.Model(&eventReviewer).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

// BatchCreate create a list of event reviewer
func (s *store) BatchCreate(db *gorm.DB, eers []model.EmployeeEventReviewer) ([]model.EmployeeEventReviewer, error) {
	return eers, db.Create(&eers).Error
}

// Create a employee event reviewer
func (s *store) Create(tx *gorm.DB, eer *model.EmployeeEventReviewer) (*model.EmployeeEventReviewer, error) {
	return eer, tx.Create(&eer).Error
}

// DeleteByEventID delete EmployeeEventReviewer by eventID
func (s *store) DeleteByEventID(db *gorm.DB, eventID string) error {
	return db.Where("event_id = ?", eventID).Delete(&model.EmployeeEventReviewer{}).Error
}

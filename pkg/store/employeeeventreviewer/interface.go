package employeeeventreviewer

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, reviewerID string, topicID string) (employeeEventReviewer *model.EmployeeEventReviewer, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.EmployeeEventReviewer, updatedFields ...string) (employeeEventReviewer *model.EmployeeEventReviewer, err error)
	BatchCreate(db *gorm.DB, employeeEventReviewers []model.EmployeeEventReviewer) ([]model.EmployeeEventReviewer, error)
	Create(tx *gorm.DB, eventReviewer *model.EmployeeEventReviewer) (employeeEventReviewer *model.EmployeeEventReviewer, err error)
	DeleteByEventID(db *gorm.DB, eventID string) error
}

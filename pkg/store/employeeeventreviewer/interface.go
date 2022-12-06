package employeeeventreviewer

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, reviewerID string, topicID string) (employeeeventreviewer *model.EmployeeEventReviewer, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.EmployeeEventReviewer, updatedFields ...string) (employeeEventReviewer *model.EmployeeEventReviewer, err error)
}

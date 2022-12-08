package employeeeventreviewer

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, reviewerID string, topicID string) (employeeeventreviewer *model.EmployeeEventReviewer, err error)
}

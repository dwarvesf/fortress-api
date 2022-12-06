package employeeeventquestion

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	GetByEventReviewerID(db *gorm.DB, reviewID string) (eventQuestions []*model.EmployeeEventQuestion, err error)
	UpdateAnswers(db *gorm.DB, data BasicEventQuestion) (err error)
}

package employeeeventquestion

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	GetByEventReviewerID(db *gorm.DB, reviewID string) (eventQuestions []*model.EmployeeEventQuestion, err error)
	UpdateAnswers(db *gorm.DB, data BasicEventQuestion) (err error)
	BatchCreate(db *gorm.DB, employeeEventQuestions []model.EmployeeEventQuestion) ([]model.EmployeeEventQuestion, error)
	Create(tx *gorm.DB, eventQuestion *model.EmployeeEventQuestion) (employeeEventQuestion *model.EmployeeEventQuestion, err error)
	DeleteByEventID(db *gorm.DB, eventID string) error
	DeleteByEventReviewerIDList(db *gorm.DB, reviewerIDList []string) error
	DeleteByEventReviewerID(db *gorm.DB, eventReviewerID string) error
}

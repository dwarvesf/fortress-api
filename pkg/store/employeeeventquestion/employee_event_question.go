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

// UpdateAnswers update answer and note by table id
func (s *store) UpdateAnswers(db *gorm.DB, data BasicEventQuestion) error {
	return db.Table("employee_event_questions").
		Where("id = ?", data.EventQuestionID).
		Updates(map[string]interface{}{"answer": data.Answer, "note": data.Note}).Error
}

// Create create new one
func (s *store) BatchCreate(db *gorm.DB, employeeEventQuestions []model.EmployeeEventQuestion) ([]model.EmployeeEventQuestion, error) {
	return employeeEventQuestions, db.Create(&employeeEventQuestions).Error
}

// Create a employee event question
func (s *store) Create(tx *gorm.DB, eventQuestion *model.EmployeeEventQuestion) (*model.EmployeeEventQuestion, error) {
	return eventQuestion, tx.Create(&eventQuestion).Error
}

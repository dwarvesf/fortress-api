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

// DeleteByEventID delete EmployeeEventQuestion by eventID
func (s *store) DeleteByEventID(db *gorm.DB, eventID string) error {
	return db.Where("event_id = ?", eventID).Delete(&model.EmployeeEventQuestion{}).Error
}

// DeleteByEventReviewerIDList delete EmployeeEventQuestion by reviewerID list
func (s *store) DeleteByEventReviewerIDList(db *gorm.DB, reviewerIDList []string) error {
	return db.Where("employee_event_reviewer_id IN ?", reviewerIDList).Delete(&model.EmployeeEventQuestion{}).Error
}

// DeleteByEventReviewerID delete EmployeeEventQuestion by eventReviewerID
func (s *store) DeleteByEventReviewerID(db *gorm.DB, eventReviewerID string) error {
	return db.Where("employee_event_reviewer_id = ?", eventReviewerID).Delete(&model.EmployeeEventQuestion{}).Error
}

// CountLikertScaleByEventIDAndDomain return LikertScaleCount by eventID and domain
func (s *store) CountLikertScaleByEventIDAndDomain(db *gorm.DB, eventID string, domain string) (*model.LikertScaleCount, error) {
	var count *model.LikertScaleCount

	query := db.Raw(`
	WITH q0 AS (
			SELECT *
			FROM employee_event_questions
			WHERE event_id = ? AND domain = ? AND deleted_at IS NULL
		),
		q1 AS (
			SELECT COUNT(*) AS "strongly_disagree"
			FROM q0
			WHERE q0.answer = ?
		), 
		q2 AS (
			SELECT COUNT(*) AS disagree
			FROM q0
			WHERE q0.answer = ?
		), 
		q3 AS (
			SELECT COUNT(*) AS mixed
			FROM q0
			WHERE q0.answer = ?
			), 
		q4 AS (
			SELECT COUNT(*) AS agree
			FROM q0
			WHERE q0.answer = ?
		), 
		q5 AS (
			SELECT COUNT(*) AS "strongly_agree"
			FROM q0
			WHERE q0.answer = ?
		)
		SELECT *
		FROM q1, q2, q3, q4, q5
	`,
		eventID,
		domain,
		model.AgreementLevelMap[model.AgreementLevelStronglyDisagree],
		model.AgreementLevelMap[model.AgreementLevelDisagree],
		model.AgreementLevelMap[model.AgreementLevelMixed],
		model.AgreementLevelMap[model.AgreementLevelAgree],
		model.AgreementLevelMap[model.AgreementLevelStronglyAgree])

	return count, query.Scan(&count).Error
}

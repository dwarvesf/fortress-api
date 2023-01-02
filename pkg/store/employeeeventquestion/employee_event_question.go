package employeeeventquestion

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}

// GetByEventReviewerID return list EmployeeEventQuestion by eventReviewerID
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

// CountLikertScaleByEventIDAndDomain return LikertScaleCount by eventID and domain
func (s *store) GetAverageAnswerEngagementByTime(db *gorm.DB, times []time.Time) ([]*model.StatisticEngagementDashboard, error) {
	var result []*model.StatisticEngagementDashboard

	query := db.Table("feedback_events fe").
		Select("DISTINCT q.content, eq.question_id, fe.title, fe.start_date, avg( CASE WHEN answer = '' THEN 0 ELSE cast(answer AS DECIMAL) END) AS point").
		Joins("JOIN employee_event_questions eq ON fe.id = eq.event_id").
		Joins("JOIN employee_event_reviewers er ON eq.employee_event_reviewer_id = er.id").
		Joins("JOIN questions q ON eq.question_id = q.id").
		Where("eq.domain = 'engagement'").
		Where("er.reviewer_status = 'done' AND is_forced_done = FALSE").
		Where("fe.start_date IN ?", times).
		Group("q.content, eq.question_id, fe.title, fe.start_date").
		Order("q.content asc")

	return result, query.Find(&result).Error
}

// CountLikertScaleByEventIDAndDomain return LikertScaleCount by eventID and domain
func (s *store) GetAverageAnswerEngagementByFilter(db *gorm.DB, filter model.EngagementDashboardFilter, time *time.Time) ([]*model.StatisticEngagementDashboard, error) {
	var result []*model.StatisticEngagementDashboard

	query := db.Table("feedback_events fe").
		Select("DISTINCT f.name, eq.question_id, fe.title, fe.start_date, avg( CASE WHEN answer = '' THEN 0 ELSE cast(answer AS DECIMAL) END) AS point").
		Joins("JOIN employee_event_questions eq ON fe.id = eq.event_id").
		Joins("JOIN employee_event_reviewers er ON eq.employee_event_reviewer_id = er.id").
		Where("eq.domain = 'engagement'").
		Where("er.reviewer_status = 'done' AND is_forced_done = FALSE").
		Where("fe.start_date = ?", time).
		Group("eq.question_id, fe.title, fe.start_date")

	if filter == model.EngagementDashboardFilterChapter {
		query = query.
			Joins("JOIN employee_chapters ec ON er.reviewer_id = ec.employee_id").
			Joins("JOIN chapters f ON ec.chapter_id = f.id").
			Group("f.name")
	}
	if filter == model.EngagementDashboardFilterSeniority {
		query = query.
			Joins("JOIN employees e ON er.reviewer_id = e.id").
			Joins("JOIN seniorities f ON e.seniority_id = f.id").
			Group("f.name")
	}
	if filter == model.EngagementDashboardFilterProject {
		query = query.
			Joins("JOIN project_members pm ON er.reviewer_id = pm.employee_id").
			Joins("JOIN projects f ON pm.project_id = f.id").
			Group("f.name")
	}

	return result, query.Find(&result).Error
}

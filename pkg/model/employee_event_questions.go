package model

import "time"

// EmployeeEventQuestion model for employee_event_questions table
type EmployeeEventQuestion struct {
	BaseModel

	EmployeeEventReviewerID UUID
	QuestionID              UUID
	EventID                 UUID
	Content                 string
	Answer                  string
	Note                    string
	Type                    string
	Order                   int64
	Domain                  QuestionDomain
}

type StatisticEngagementDashboard struct {
	Name       string
	Content    string
	Title      string
	Point      float64
	QuestionID UUID
	StartDate  time.Time
}

type EngagementDashboardFilter string

const (
	EngagementDashboardFilterDepartment EngagementDashboardFilter = "department"
	EngagementDashboardFilterChapter    EngagementDashboardFilter = "chapter"
	EngagementDashboardFilterSeniority  EngagementDashboardFilter = "seniority"
	EngagementDashboardFilterProject    EngagementDashboardFilter = "project"
)

// String returns the string type from the EngagementDashboardFilter type
func (e EngagementDashboardFilter) String() string {
	return string(e)
}

func (e EngagementDashboardFilter) IsValid() bool {
	switch e {
	case
		EngagementDashboardFilterDepartment,
		EngagementDashboardFilterChapter,
		EngagementDashboardFilterSeniority,
		EngagementDashboardFilterProject:
		return true
	}
	return false
}

// ToQuestionMap create map from EmployeeEventQuestion
func ToQuestionMap(questionList []*EmployeeEventQuestion) map[UUID]string {
	rs := map[UUID]string{}
	for _, q := range questionList {
		rs[q.ID] = q.Answer
	}

	return rs
}

// ToQuestionMapType create map from Question to type
func ToQuestionMapType(questionList []*EmployeeEventQuestion) map[UUID]string {
	rs := map[UUID]string{}
	for _, q := range questionList {
		rs[q.ID] = q.Type
	}

	return rs
}

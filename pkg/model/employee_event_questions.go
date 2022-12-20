package model

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

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
}

// ToQuestionMap create map from EmployeeEventQuestion
func ToQuestionMap(questionList []*EmployeeEventQuestion) map[UUID]string {
	rs := map[UUID]string{}
	for _, q := range questionList {
		rs[q.ID] = q.Answer
	}

	return rs
}

type LikertScaleCount struct {
	StronglyDisagree int `json:"strongly-disagree"`
	Disagree         int `json:"disagree"`
	Mixed            int `json:"mixed"`
	Agree            int `json:"agree"`
	StronglyAgree    int `json:"strongly-agree"`
}

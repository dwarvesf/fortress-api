package model

// QuestionType type of question
type QuestionType string

// valid values for QuestionType
const (
	QuestionTypeScale   QuestionType = "likert-scale"
	QuestionTypeGeneral QuestionType = "general"
)

// IsValid validation for QuestionType
func (e QuestionType) IsValid() bool {
	switch e {
	case
		QuestionTypeScale,
		QuestionTypeGeneral:
		return true
	}
	return false
}

// String returns a string representation of QuestionType
func (e QuestionType) String() string {
	return string(e)
}

// Question model for questions table
type Question struct {
	BaseModel

	Type        QuestionType
	Category    EventType
	Subcategory EventSubtype
	Content     string
	Order       int64
	Domain      QuestionDomain
	EventID     UUID
}

// LikertScaleAnswer type of question
type LikertScaleAnswer string

// valid values for LikertScaleAnswer
const (
	LikertScaleAnswerStronglyDisagree LikertScaleAnswer = "strongly-disagree"
	LikertScaleAnswerDisagree         LikertScaleAnswer = "disagree"
	LikertScaleAnswerMixed            LikertScaleAnswer = "mixed"
	LikertScaleAnswerAgree            LikertScaleAnswer = "agree"
	LikertScaleAnswerStronglyAgree    LikertScaleAnswer = "strongly-agree"
)

// IsValid validation for LikertScaleAnswer
func (e LikertScaleAnswer) IsValid() bool {
	switch e {
	case
		LikertScaleAnswerStronglyDisagree,
		LikertScaleAnswerDisagree,
		LikertScaleAnswerMixed,
		LikertScaleAnswerAgree,
		LikertScaleAnswerStronglyAgree:
		return true
	}
	return false
}

// String returns a string representation of LikertScaleAnswer
func (e LikertScaleAnswer) String() string {
	return string(e)
}

// QuestionDomain type for domain of questions table
type QuestionDomain string

// QuestionDomain values
const (
	QuestionDomainWorkload QuestionDomain = "workload"
	QuestionDomainDeadline QuestionDomain = "deadline"
	QuestionDomainLearning QuestionDomain = "learning"
)

// IsValid validation for QuestionDomain
func (e QuestionDomain) IsValid() bool {
	switch e {
	case
		QuestionDomainWorkload,
		QuestionDomainDeadline,
		QuestionDomainLearning:
		return true
	}
	return false
}

// String returns the string type from the QuestionDomain type
func (e QuestionDomain) String() string {
	return string(e)
}

// LikertScaleCount represent for counted likert-scale answer model
type LikertScaleCount struct {
	StronglyDisagree int `json:"strongly-disagree"`
	Disagree         int `json:"disagree"`
	Mixed            int `json:"mixed"`
	Agree            int `json:"agree"`
	StronglyAgree    int `json:"strongly-agree"`
}

type QuestionDomainCount struct {
	Domain QuestionDomain
	LikertScaleCount
}

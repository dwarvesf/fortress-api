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
	EventID     UUID
}

// LikertScaleAnswers type of question
type LikertScaleAnswers string

// valid values for LikertScaleAnswers
const (
	LikertScaleAnswersSDisagree LikertScaleAnswers = "strongly-disagree"
	LikertScaleAnswersDisagree  LikertScaleAnswers = "disagree"
	LikertScaleAnswersMixed     LikertScaleAnswers = "mixed"
	LikertScaleAnswersAgree     LikertScaleAnswers = "agree"
	LikertScaleAnswersSAgree    LikertScaleAnswers = "strongly-agree"
)

// IsValid validation for LikertScaleAnswers
func (e LikertScaleAnswers) IsValid() bool {
	switch e {
	case
		LikertScaleAnswersSDisagree,
		LikertScaleAnswersDisagree,
		LikertScaleAnswersMixed,
		LikertScaleAnswersAgree,
		LikertScaleAnswersSAgree:
		return true
	}
	return false
}

// String returns a string representation of LikertScaleAnswers
func (e LikertScaleAnswers) String() string {
	return string(e)
}

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

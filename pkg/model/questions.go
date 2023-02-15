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
	QuestionDomainEngagement QuestionDomain = "engagement"
	QuestionDomainWorkload   QuestionDomain = "workload"
	QuestionDomainDeadline   QuestionDomain = "deadline"
	QuestionDomainLearning   QuestionDomain = "learning"
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
	StronglyDisagree int `json:"stronglyDisagree" gorm:"column:strongly_disagree"`
	Disagree         int `json:"disagree" gorm:"column:disagree"`
	Mixed            int `json:"mixed" gorm:"column:mixed"`
	Agree            int `json:"agree" gorm:"column:agree"`
	StronglyAgree    int `json:"stronglyAgree" gorm:"column:strongly_agree"`
}

type QuestionDomainCount struct {
	Domain QuestionDomain
	LikertScaleCount
}

// AgreementLevel type for work survey answer
type AgreementLevel string

// AgreementLevel values
const (
	AgreementLevelStronglyDisagree AgreementLevel = "strongly-disagree"
	AgreementLevelDisagree         AgreementLevel = "disagree"
	AgreementLevelMixed            AgreementLevel = "mixed"
	AgreementLevelAgree            AgreementLevel = "agree"
	AgreementLevelStronglyAgree    AgreementLevel = "strongly-agree"
)

// IsValid validation for AgreementLevel
func (e AgreementLevel) IsValid() bool {
	switch e {
	case
		AgreementLevelStronglyDisagree,
		AgreementLevelDisagree,
		AgreementLevelMixed,
		AgreementLevelAgree,
		AgreementLevelStronglyAgree:
		return true
	}
	return false
}

// String returns the string type of AgreementLevel
func (e AgreementLevel) String() string {
	return string(e)
}

var AgreementLevelMap = map[AgreementLevel]string{
	AgreementLevelStronglyDisagree: "1",
	AgreementLevelDisagree:         "2",
	AgreementLevelMixed:            "3",
	AgreementLevelAgree:            "4",
	AgreementLevelStronglyAgree:    "5",
}

var AgreementLevelValueMap = map[string]AgreementLevel{
	"1": AgreementLevelStronglyDisagree,
	"2": AgreementLevelDisagree,
	"3": AgreementLevelMixed,
	"4": AgreementLevelAgree,
	"5": AgreementLevelStronglyAgree,
}

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
	// engagement
	AgreementLevelStronglyDisagree AgreementLevel = "strongly-disagree"
	AgreementLevelDisagree         AgreementLevel = "disagree"
	AgreementLevelMixed            AgreementLevel = "mixed"
	AgreementLevelAgree            AgreementLevel = "agree"
	AgreementLevelStronglyAgree    AgreementLevel = "strongly-agree"

	// work - workload domain
	AgreementLevelBreeze          AgreementLevel = "breeze"
	AgreementLevelCouldHandleMore AgreementLevel = "could-handle-more"
	AgreementLevelManageable      AgreementLevel = "manageable"
	AgreementLevelOverwhelming    AgreementLevel = "overwhelming"
	AgreementLevelCantKeepUp      AgreementLevel = "cant-keep-up"

	// work - deadline domain
	AgreementLevelVeryUncertain     AgreementLevel = "very-uncertain"
	AgreementLevelSomewhatUncertain AgreementLevel = "somewhat-uncertain"
	AgreementLevelNeutral           AgreementLevel = "neutral"
	AgreementLevelSomewhatConfident AgreementLevel = "somewhat-confident"
	AgreementLevelVeryConfident     AgreementLevel = "very-confident"

	// work - learning domain
	AgreementLevelVeryLittle AgreementLevel = "very-little"
	AgreementLevelSomewhat   AgreementLevel = "somewhat"
	AgreementLevelModerate   AgreementLevel = "moderate"
	AgreementLevelALot       AgreementLevel = "a-lot"
	AgreementLevelExtremely  AgreementLevel = "extremely"
)

// IsValid validation for AgreementLevel
func (e AgreementLevel) IsValid() bool {
	switch e {
	case
		AgreementLevelStronglyDisagree,
		AgreementLevelDisagree,
		AgreementLevelMixed,
		AgreementLevelAgree,
		AgreementLevelStronglyAgree,
		AgreementLevelBreeze,
		AgreementLevelCouldHandleMore,
		AgreementLevelManageable,
		AgreementLevelOverwhelming,
		AgreementLevelCantKeepUp,
		AgreementLevelVeryUncertain,
		AgreementLevelSomewhatUncertain,
		AgreementLevelNeutral,
		AgreementLevelSomewhatConfident,
		AgreementLevelVeryConfident,
		AgreementLevelVeryLittle,
		AgreementLevelSomewhat,
		AgreementLevelModerate,
		AgreementLevelALot,
		AgreementLevelExtremely:
		return true
	}
	return false
}

// String returns the string type of AgreementLevel
func (e AgreementLevel) String() string {
	return string(e)
}

var AgreementLevelMap = map[AgreementLevel]string{
	// engagement
	AgreementLevelStronglyDisagree: "1",
	AgreementLevelDisagree:         "2",
	AgreementLevelMixed:            "3",
	AgreementLevelAgree:            "4",
	AgreementLevelStronglyAgree:    "5",

	// workload
	AgreementLevelBreeze:          "1",
	AgreementLevelCouldHandleMore: "2",
	AgreementLevelManageable:      "3",
	AgreementLevelOverwhelming:    "4",
	AgreementLevelCantKeepUp:      "5",

	// deadline
	AgreementLevelVeryUncertain:     "1",
	AgreementLevelSomewhatUncertain: "2",
	AgreementLevelNeutral:           "3",
	AgreementLevelSomewhatConfident: "4",
	AgreementLevelVeryConfident:     "5",

	// learning
	AgreementLevelVeryLittle: "1",
	AgreementLevelSomewhat:   "2",
	AgreementLevelModerate:   "3",
	AgreementLevelALot:       "4",
	AgreementLevelExtremely:  "5",
}

var AgreementLevelValueMap = map[QuestionDomain]map[string]AgreementLevel{
	QuestionDomainEngagement: {
		"1": AgreementLevelStronglyDisagree,
		"2": AgreementLevelDisagree,
		"3": AgreementLevelMixed,
		"4": AgreementLevelAgree,
		"5": AgreementLevelStronglyAgree,
	},
	QuestionDomainWorkload: {
		"1": AgreementLevelBreeze,
		"2": AgreementLevelCouldHandleMore,
		"3": AgreementLevelManageable,
		"4": AgreementLevelOverwhelming,
		"5": AgreementLevelCantKeepUp,
	},
	QuestionDomainDeadline: {
		"1": AgreementLevelVeryUncertain,
		"2": AgreementLevelSomewhatUncertain,
		"3": AgreementLevelNeutral,
		"4": AgreementLevelSomewhatConfident,
		"5": AgreementLevelVeryConfident,
	},
	QuestionDomainLearning: {
		"1": AgreementLevelVeryLittle,
		"2": AgreementLevelSomewhat,
		"3": AgreementLevelModerate,
		"4": AgreementLevelALot,
		"5": AgreementLevelExtremely,
	},
}

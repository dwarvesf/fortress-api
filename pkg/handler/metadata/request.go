package metadata

// GetQuestionsInput input params for get questions api
type GetQuestionsInput struct {
	Category    EventType    `form:"category" json:"category" binding:"required"`
	Subcategory EventSubtype `form:"subcategory" json:"subcategory" binding:"required"`
}

type EventType string   // @name EventType
type EventStatus string // @name EventStatus

const (
	EventTypeFeedback EventType = "feedback"
	EventTypeSurvey   EventType = "survey"
)

// IsValid validation for EventType
func (e EventType) IsValid() bool {
	switch e {
	case
		EventTypeFeedback,
		EventTypeSurvey:
		return true
	}
	return false
}

func (e EventType) String() string {
	return string(e)
}

type EventSubtype string

const (
	EventSubtypePeerReview   EventSubtype = "peer-review"
	EventSubtypeEngagement   EventSubtype = "engagement"
	EventSubtypeWork         EventSubtype = "work"
	EventSubtypeAppreciation EventSubtype = "appreciation"
	EventSubtypeComment      EventSubtype = "comment"
)

func (e EventSubtype) IsValid() bool {
	switch e {
	case
		EventSubtypePeerReview,
		EventSubtypeEngagement,
		EventSubtypeWork,
		EventSubtypeAppreciation,
		EventSubtypeComment:
		return true
	}
	return false
}

// Validate check valid for values in input params
func (i GetQuestionsInput) Validate() error {
	if i.Category == "" || !i.Category.IsValid() {
		return ErrInvalidCategory
	}

	if i.Subcategory == "" || !i.Subcategory.IsValid() {
		return ErrInvalidSubcategory
	}

	return nil
}

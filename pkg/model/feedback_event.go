package model

import "time"

// EventType event_type for table feedback events
type EventType string

// values for EventType
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

// String returns the string type from the EventType type
func (e EventType) String() string {
	return string(e)
}

// EventSubtype event_subtype for table feedback events
type EventSubtype string

// values for EventSubtype
const (
	EventSubtypePeerReview   EventSubtype = "peer-review"
	EventSubtypeEngagement   EventSubtype = "engagement"
	EventSubtypeWork         EventSubtype = "work"
	EventSubtypeAppreciation EventSubtype = "appreciation"
	EventSubtypeComment      EventSubtype = "comment"
)

// IsValid validation for EventSubtype
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

// String returns the string type from the EventSubtype type
func (e EventSubtype) String() string {
	return string(e)
}

// FeedbackEvent model for feedback_events table
type FeedbackEvent struct {
	BaseModel

	Title     string
	Type      EventType
	Subtype   EventSubtype
	Status    string
	CreatedBy UUID
	StartDate *time.Time
	EndDate   *time.Time

	Employee Employee `gorm:"foreignKey:CreatedBy"`
}

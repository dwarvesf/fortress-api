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

// EventSubType event_subtype for table feedback events
type EventSubType string

// values for EventSubType
const (
	EventSubTypePeerReview   EventSubType = "peer-review"
	EventSubTypeEngagement   EventSubType = "engagement"
	EventSubTypeWork         EventSubType = "work"
	EventSubTypeAppreciation EventSubType = "appreciation"
	EventSubTypeComment      EventSubType = "comment"
)

// IsValid validation for EventSubType
func (e EventSubType) IsValid() bool {
	switch e {
	case
		EventSubTypePeerReview,
		EventSubTypeEngagement,
		EventSubTypeWork,
		EventSubTypeAppreciation,
		EventSubTypeComment:
		return true
	}
	return false
}

// FeedbackEvent model for feedback_events table
type FeedbackEvent struct {
	BaseModel

	Title     string
	Type      EventType
	SubType   EventSubType
	Status    string
	CreatedBy UUID
	StartDate *time.Time
	EndDate   *time.Time
}

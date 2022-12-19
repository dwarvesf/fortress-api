package model

// EmployeeEventReviewer model for table employee_event_reviewer
type EmployeeEventReviewer struct {
	BaseModel

	EmployeeEventTopicID UUID
	ReviewerID           UUID
	ReviewerStatus       EventReviewerStatus
	AuthorStatus         EventAuthorStatus
	Relationship         Relationship
	IsShared             bool
	IsRead               bool
	IsForcedDone         bool
	EventID              UUID

	Reviewer               *Employee
	EmployeeEventQuestions []EmployeeEventQuestion
}

// EventReviewerStatus event_reviewer_status for table employee event reviewer
type EventReviewerStatus string

// EventReviewerStatus values
const (
	EventReviewerStatusNone  EventReviewerStatus = "none"
	EventReviewerStatusNew   EventReviewerStatus = "new"
	EventReviewerStatusDraft EventReviewerStatus = "draft"
	EventReviewerStatusDone  EventReviewerStatus = "done"
)

// IsValid validation for EventReviewerStatus
func (e EventReviewerStatus) IsValid() bool {
	switch e {
	case
		EventReviewerStatusNone,
		EventReviewerStatusDraft,
		EventReviewerStatusDone,
		EventReviewerStatusNew:
		return true
	}
	return false
}

// String returns the string type from the EventReviewerStatus type
func (e EventReviewerStatus) String() string {
	return string(e)
}

// EventAuthorStatus event_reviewer_status for table employee event reviewer
type EventAuthorStatus string

// EventAuthorStatus values
const (
	EventAuthorStatusDraft EventAuthorStatus = "draft"
	EventAuthorStatusSent  EventAuthorStatus = "sent"
	EventAuthorStatusDone  EventAuthorStatus = "done"
)

// IsValid validation for EventAuthorStatus
func (e EventAuthorStatus) IsValid() bool {
	switch e {
	case
		EventAuthorStatusDraft,
		EventAuthorStatusSent,
		EventAuthorStatusDone:
		return true
	}
	return false
}

// String returns the string type from the EventReviewerStatus type
func (e EventAuthorStatus) String() string {
	return string(e)
}

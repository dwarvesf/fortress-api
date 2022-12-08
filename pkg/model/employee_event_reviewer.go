package model

// EmployeeEventReviewer model for table employee_event_reviewer
type EmployeeEventReviewer struct {
	BaseModel

	EmployeeEventTopicID UUID
	ReviewerID           UUID
	Status               EventReviewerStatus
	Relationship         Relationship
	IsShared             bool
	IsRead               bool
	EventID              UUID

	Reviewer *Employee
}

// EventReviewerStatus event_reviewer_status for table employee event reviewer
type EventReviewerStatus string

// EventReviewerStatus values
const (
	EventReviewerStatusDraft EventReviewerStatus = "draft"
	EventReviewerStatusDone  EventReviewerStatus = "done"
)

// IsValid validation for EventReviewerStatus
func (e EventReviewerStatus) IsValid() bool {
	switch e {
	case
		EventReviewerStatusDraft,
		EventReviewerStatusDone:
		return true
	}
	return false
}

// String returns the string type from the EventReviewerStatus type
func (e EventReviewerStatus) String() string {
	return string(e)
}
